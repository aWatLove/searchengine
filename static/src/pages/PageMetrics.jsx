
import { useState, useEffect } from 'react';
import { useAppContext } from '../contexts/AppContext';
import api from '../services/api';

export default function PageMetrics() {
    const { startLoading, stopLoading } = useAppContext();
    const [metrics, setMetrics] = useState({
        cpuUsage: null,
        memoryUsage: null,
        requestRate: null,
        errorRate: null,
    });
    const [timeRange, setTimeRange] = useState('1h');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    // Загрузка метрик
    const loadMetrics = async () => {
        try {
            startLoading();
            setLoading(true);

            // Пример запросов к API для получения метрик
            const responses = await Promise.all([
                api.getMetrics('cpu_usage', timeRange),
                api.getMetrics('memory_usage', timeRange),
                api.getMetrics('http_requests_rate', timeRange),
                api.getMetrics('http_errors_rate', timeRange),
            ]);

            setMetrics({
                cpuUsage: responses[0].data,
                memoryUsage: responses[1].data,
                requestRate: responses[2].data,
                errorRate: responses[3].data,
            });

        } catch (err) {
            setError(err.message);
            console.error('Metrics load error:', err);
        } finally {
            stopLoading();
            setLoading(false);
        }
    };

    useEffect(() => {
        loadMetrics();
    }, [timeRange]);

    // Рендер метрики
    const renderMetricCard = (title, data, unit = '') => {
        if (loading) return <div className="p-4 bg-gray-100 rounded animate-pulse">Loading...</div>;
        if (!data) return null;

        return (
            <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                <h3 className="text-lg font-semibold mb-4">{title}</h3>
                <div className="flex items-baseline gap-2">
          <span className="text-3xl font-bold text-blue-600">
            {data.currentValue}
          </span>
                    <span className="text-gray-500">{unit}</span>
                </div>
                <div className="mt-4 h-32">
                    {/* Здесь можно вставить график */}
                    <div className="bg-gray-50 w-full h-full rounded-lg flex items-center justify-center text-gray-400">
                        Chart placeholder
                    </div>
                </div>
            </div>
        );
    };

    return (
        <div className="p-8 max-w-7xl mx-auto">
            <div className="mb-6 flex justify-between items-center">
                <h1 className="text-2xl font-bold">Системные метрики</h1>
                <select
                    value={timeRange}
                    onChange={(e) => setTimeRange(e.target.value)}
                    className="px-4 py-2 border rounded-lg"
                >
                    <option value="15m">15 минут</option>
                    <option value="1h">1 час</option>
                    <option value="24h">24 часа</option>
                    <option value="7d">7 дней</option>
                </select>
            </div>

            {error && (
                <div className="bg-red-50 text-red-700 p-4 rounded-lg mb-6">
                    Ошибка загрузки метрик: {error}
                </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                {renderMetricCard('CPU Usage', metrics.cpuUsage, '%')}
                {renderMetricCard('Memory Usage', metrics.memoryUsage, 'GB')}
                {renderMetricCard('Request Rate', metrics.requestRate, 'rps')}
                {renderMetricCard('Error Rate', metrics.errorRate, '%')}
            </div>

            {/* Дополнительные графики */}
            <div className="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                    <h3 className="text-lg font-semibold mb-4">История нагрузки CPU</h3>
                    <div className="h-64 bg-gray-50 rounded-lg flex items-center justify-center text-gray-400">
                        Time series chart
                    </div>
                </div>
                <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                    <h3 className="text-lg font-semibold mb-4">Распределение ошибок</h3>
                    <div className="h-64 bg-gray-50 rounded-lg flex items-center justify-center text-gray-400">
                        Pie chart
                    </div>
                </div>
            </div>
        </div>
    );
}