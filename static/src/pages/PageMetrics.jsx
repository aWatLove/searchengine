import React, { useState, useEffect, useRef, useMemo } from 'react';
import { useAppContext } from '../contexts/AppContext';
import api from '../services/api';
import {
    LineChart,
    Line,
    CartesianGrid,
    XAxis,
    YAxis,
    Tooltip,
    ResponsiveContainer,
    Legend,
} from 'recharts';

// Парсер Prometheus-метрик
function parsePrometheus(text) {
    const lines = text.trim().split(/\r?\n/);
    const result = {
        cpu: null,
        memory: null,
        requests: {},
        errors: {},
        durations: {},
        totals: { req: null, err: null, durSum: null, durCount: null },
    };

    lines.forEach((ln) => {
        if (!ln || ln.startsWith('#')) return;
        const [metricPart, rawValue] = ln.split(/\s+/);
        const value = parseFloat(rawValue);

        if (metricPart === 'go_app_cpu_usage_percent') {
            result.cpu = value;
            return;
        }
        if (metricPart === 'go_app_ram_bytes') {
            result.memory = Math.round((value / 1024 ** 3) * 100) / 100;
            return;
        }

        const vecMatch = metricPart.match(/^([a-z0-9_]+)(?:\{(.+)\})?$/);
        if (!vecMatch) return;
        const [_, name, labels] = vecMatch;
        const lbls = {};
        if (labels) {
            labels.split(',').forEach((kv) => {
                const [k, v] = kv.split('=');
                lbls[k] = v.replace(/^"|"$/g, '');
            });
        }

        switch (name) {
            case 'http_requests_total': {
                if (lbls.handler) {
                    const key = `${lbls.handler}|${lbls.method}|${lbls.status}`;
                    result.requests[key] = { handler: lbls.handler, method: lbls.method, status: lbls.status, value };
                } else {
                    result.totals.req = value;
                }
                break;
            }
            case 'http_errors_total': {
                if (lbls.handler) {
                    const key = `${lbls.handler}|${lbls.method}`;
                    result.errors[key] = { handler: lbls.handler, method: lbls.method, value };
                } else {
                    result.totals.err = value;
                }
                break;
            }
            case 'http_request_duration_seconds_sum': {
                if (lbls.handler) {
                    const key = `${lbls.handler}|${lbls.method}`;
                    result.durations[key] = result.durations[key] || {};
                    result.durations[key].sum = value;
                } else {
                    result.totals.durSum = value;
                }
                break;
            }
            case 'http_request_duration_seconds_count': {
                if (lbls.handler) {
                    const key = `${lbls.handler}|${lbls.method}`;
                    result.durations[key] = result.durations[key] || {};
                    result.durations[key].count = value;
                } else {
                    result.totals.durCount = value;
                }
                break;
            }
            default:
                break;
        }
    });

    if (result.totals.req == null) {
        result.totals.req = Object.values(result.requests).reduce((sum, o) => sum + o.value, 0);
    }

    return result;
}

// Вычисление дельт RPS, Latency и Error Rate
function computeMetrics(prev, curr) {
    const dt = 5;
    let totalRPS = 0;
    const rpsList = [];
    const latencyList = [];
    const errList = [];

    if (prev.totals.req != null) {
        totalRPS = (curr.totals.req - prev.totals.req) / dt;
    }

    // По handler/method/status
    Object.values(curr.requests).forEach(({ handler, method, status, value }) => {
        const key = `${handler}|${method}|${status}`;
        const prevVal = prev.requests[key]?.value;
        if (prevVal != null) {
            const rps = Math.max(0, (value - prevVal) / dt);
            rpsList.push({ handler, method, status, rps: Math.round(rps * 100) / 100 });
        }
    });

    // Latency
    Object.entries(curr.durations).forEach(([key, { sum, count }]) => {
        const pd = prev.durations[key] || {};
        if (pd.sum != null && pd.count != null && sum != null && count != null) {
            const latency = (sum - pd.sum) / (count - pd.count) || 0;
            const [handler, method] = key.split('|');
            latencyList.push({ handler, method, latency: Math.round(latency * 1000) / 1000 });
        }
    });

    // Error Rate
    rpsList
        .filter((e) => e.status.startsWith('4') || e.status.startsWith('5'))
        .forEach(({ handler, method, status }) => {
            const errKey = `${handler}|${method}`;
            const prevErr = prev.errors[errKey]?.value || 0;
            const currErr = curr.errors[errKey]?.value || 0;
            const reqKey = `${handler}|${method}|${status}`;
            const deltaReq = curr.requests[reqKey]?.value - (prev.requests[reqKey]?.value || 0);
            const deltaErr = currErr - prevErr;
            const rate = deltaReq > 0 ? (deltaErr / deltaReq) * 100 : 0;
            errList.push({ handler, method, errorRate: Math.round(rate * 100) / 100 });
        });

    return { totalRPS: Math.round(totalRPS * 100) / 100, rpsList, latencyList, errList };
}

export default function PageMetrics() {
    const { startLoading, stopLoading } = useAppContext();
    const [history, setHistory] = useState([]);
    const [metrics, setMetrics] = useState(null);
    const [computed, setComputed] = useState({ totalRPS: 0, rpsList: [], latencyList: [], errList: [] });
    const [timeRange, setTimeRange] = useState('1h');
    const [error, setError] = useState(null);
    const intervalRef = useRef(null);

    const maxPoints = useMemo(() => ({ '15m': 180, '1h': 720, '24h': 2880, '7d': 20160 }[timeRange] || 720), [timeRange]);
    const data = history.slice(-maxPoints);

    const loadMetrics = async () => {
        startLoading();
        try {
            const resp = await api.getRawMetrics();
            const parsed = parsePrometheus(resp.data);
            let comp = { totalRPS: 0, rpsList: [], latencyList: [], errList: [] };
            setHistory((prev) => {
                const last = prev[prev.length - 1];
                if (last) comp = computeMetrics(last.parsed, parsed);
                const entry = {
                    timestamp: new Date().toLocaleTimeString(),
                    parsed,
                    cpu: parsed.cpu,
                    memory: parsed.memory,
                    rps: comp.totalRPS,
                    errorRate: comp.errList[0]?.errorRate ?? 0,
                    latency: comp.latencyList[0]?.latency ?? 0,
                };
                return [...prev.slice(-10000), entry];
            });
            setMetrics(parsed);
            setComputed(comp);
            setError(null);
        } catch (e) {
            setError(e.message);
        } finally {
            stopLoading();
        }
    };

    useEffect(() => {
        loadMetrics();
        intervalRef.current = setInterval(loadMetrics, 5000);
        return () => clearInterval(intervalRef.current);
    }, []);

    const Card = ({ title, value, unit, desc }) => (
        <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
            <div className="flex justify-between mb-2">
                <h3 className="text-lg font-semibold">{title}</h3>
                <span className="text-xs">{unit}</span>
            </div>
            <div className="text-3xl font-bold mb-2">{value}</div>
            <div className="text-xs text-gray-400">{desc}</div>
        </div>
    );

    const Chart = ({ dataKey, title, unit }) => (
        <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
            <h3 className="text-lg font-semibold mb-4">{title}</h3>
            <ResponsiveContainer width="100%" height={220}>
                <LineChart data={data}>
                    <CartesianGrid stroke="#f1f1f1" />
                    <XAxis dataKey="timestamp" />
                    <YAxis unit={unit} allowDecimals />
                    <Tooltip />
                    <Legend />
                    <Line type="monotone" dataKey={dataKey} dot={false} strokeWidth={2} />
                </LineChart>
            </ResponsiveContainer>
        </div>
    );

    return (
        <div className="p-8 max-w-7xl mx-auto">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-2xl font-bold">Системные метрики</h1>
                <select value={timeRange} onChange={(e) => setTimeRange(e.target.value)} className="px-4 py-2 border rounded-lg">
                    <option value="15m">15 минут</option>
                    <option value="1h">1 час</option>
                    <option value="24h">24 часа</option>
                    <option value="7d">7 дней</option>
                </select>
            </div>
            {error && <div className="bg-red-50 text-red-700 p-4 rounded-lg mb-6">Ошибка: {error}</div>}

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6 mb-10">
                <Card title="CPU Usage" value={metrics?.cpu} unit="%" desc="Использование CPU" />
                <Card title="Memory Usage" value={metrics?.memory} unit="GB" desc="Использование RAM" />
                <Card title="RPS" value={computed.totalRPS} unit="req/s" desc="Запросов в секунду" />
                <Card title="Error Rate" value={computed.errList[0]?.errorRate ?? 0} unit="%" desc="Процент ошибок" />
                <Card title="Latency" value={computed.latencyList[0]?.latency ?? 0} unit="s" desc="Средняя задержка" />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-10">
                <Chart dataKey="cpu" title="CPU Usage" unit="%" />
                <Chart dataKey="memory" title="Memory Usage" unit="GB" />
                <Chart dataKey="rps" title="RPS" unit="req/s" />
                <Chart dataKey="errorRate" title="Error Rate" unit="%" />
                <Chart dataKey="latency" title="Latency" unit="s" />
            </div>

            <div className="mt-10 grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                    <h3 className="text-lg font-semibold mb-4">Топ-10 RPS</h3>
                    <table className="min-w-full text-sm">
                        <thead><tr><th>Handler</th><th>Method</th><th>Status</th><th>RPS</th></tr></thead>
                        <tbody>{computed.rpsList.sort((a,b)=>b.rps-a.rps).slice(0,10).map((r,i)=><tr key={i}><td>{r.handler}</td><td>{r.method}</td><td>{r.status}</td><td>{r.rps}</td></tr>)}</tbody>
                    </table>
                </div>
                <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                    <h3 className="text-lg font-semibold mb-4">Топ-10 Latency</h3>
                    <table className="min-w-full text-sm">
                        <thead><tr><th>Handler</th><th>Method</th><th>Latency</th></tr></thead>
                        <tbody>{computed.latencyList.sort((a,b)=>b.latency-a.latency).slice(0,10).map((r,i)=><tr key={i}><td>{r.handler}</td><td>{r.method}</td><td>{r.latency}</td></tr>)}</tbody>
                    </table>
                </div>
                <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                    <h3 className="text-lg font-semibold mb-4">Топ-10 Error Rate</h3>
                    <table className="min-w-full text-sm">
                        <thead><tr><th>Handler</th><th>Method</th><th>Error Rate</th></tr></thead>
                        <tbody>{computed.errList.sort((a,b)=>b.errorRate-a.errorRate).slice(0,10).map((r,i)=><tr key={i}><td>{r.handler}</td><td>{r.method}</td><td>{r.errorRate}</td></tr>)}</tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
