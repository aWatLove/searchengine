
import '../services/api'
import api from "../services/api.js";
import { useState, useEffect, useCallback } from 'react'
import { useAppContext } from '../contexts/AppContext'

export default function PageLogs() {
    const { startLoading, stopLoading } = useAppContext()
    const [logsList, setLogsList] = useState([])
    const [currentLog, setCurrentLog] = useState('')
    const [selectedLog, setSelectedLog] = useState(null)
    const [error, setError] = useState(null)

    // Загрузка последнего лога и списка файлов
    const loadInitialData = useCallback(async () => {
        try {
            startLoading()
            setError(null)

            const [listResponse, lastLogResponse] = await Promise.all([
                api.listLogs(),
                api.getLastLog()
            ])

            // Получаем массив файлов напрямую из response.data
            const files = listResponse.data
            setLogsList(files)

            // Устанавливаем первый файл только если массив не пустой
            if(files.length > 0) {
                setSelectedLog(files[0])
            }

            // Для plain/text ответа используем response.data напрямую
            setCurrentLog(lastLogResponse.data)
        } catch (error) {
            setError('Ошибка загрузки логов: ' + error.message)
            setLogsList([]) // Сбрасываем список логов при ошибке
            setSelectedLog(null)
        } finally {
            stopLoading()
        }
    }, [startLoading, stopLoading])

    useEffect(() => {
        loadInitialData()
    }, [loadInitialData])

    // Загрузка конкретного лога
    const loadLogFile = async (filename) => {
        try {
            startLoading()
            setError(null)
            const response = await api.getLog(filename)
            setCurrentLog(response.data)
            setSelectedLog(filename)
        } catch (error) {
            setError('Ошибка загрузки файла лога: ' + error.message)
            setCurrentLog('')
        } finally {
            stopLoading()
        }
    }

    // Обработка копирования в буфер обмена
    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(currentLog)
        } catch (copyError) {
            setError('Не удалось скопировать лог: ' + copyError.message)
        }
    }

    return (
        <div className="p-8 max-w-7xl mx-auto">
            <div className="grid grid-cols-1 lg:grid-cols-5 gap-8 min-h-screen">
                {/* Список логов */}
                <div className="lg:col-span-1 bg-white rounded-2xl shadow-xl p-6 border border-gray-100">
                    <div className="flex items-center justify-between mb-6">
                        <h2 className="text-2xl font-bold bg-gradient-to-r from-purple-600 to-blue-600
                            bg-clip-text text-transparent">
                            Лог-файлы
                        </h2>
                        <button
                            onClick={loadInitialData}
                            className="p-2 text-gray-500 hover:text-gray-700 transition-colors"
                            title="Обновить список"
                        >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                            </svg>
                        </button>
                    </div>

                    <div className="space-y-2 max-h-[80vh] overflow-y-auto">
                        {logsList.length === 0 ? (
                            <div className="p-3 text-gray-500 text-center">
                                Нет доступных логов
                            </div>
                        ) : (
                            logsList.map((file) => (
                                <div
                                    key={file}
                                    onClick={() => loadLogFile(file)}
                                    className={`p-3 rounded-lg cursor-pointer transition-all
                                        ${selectedLog === file
                                        ? 'bg-gradient-to-r from-blue-50 to-purple-50 border-2 border-blue-200'
                                        : 'hover:bg-gray-50'}`}
                                >
                                    <div className="flex items-center gap-3">
                                        <svg className={`w-5 h-5 shrink-0 
                                            ${selectedLog === file ? 'text-blue-500' : 'text-gray-400'}`}
                                             fill="none"
                                             stroke="currentColor"
                                             viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                                        </svg>
                                        <span className="text-sm font-medium break-words">{file}</span>
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                </div>

                {/* Контент лога */}
                <div className="lg:col-span-4 bg-white rounded-2xl shadow-xl p-6 border border-gray-100">
                    <div className="flex items-center justify-between mb-6">
                        <h2 className="text-2xl font-bold bg-gradient-to-r from-gray-800 to-gray-600
                            bg-clip-text text-transparent">
                            {selectedLog || 'Выберите лог-файл'}
                        </h2>
                        <div className="flex items-center gap-3">
                            {currentLog && (
                                <>
                                    <span className="text-sm text-gray-500">
                                        {currentLog.length} символов
                                    </span>
                                    <button
                                        onClick={handleCopy}
                                        className="p-2 text-gray-500 hover:text-blue-600 transition-colors"
                                        title="Скопировать лог"
                                    >
                                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                                  d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                                        </svg>
                                    </button>
                                </>
                            )}
                        </div>
                    </div>

                    {error ? (
                        <div className="p-6 bg-red-50 rounded-xl border-2 border-red-100">
                            <p className="text-red-600 font-medium flex items-center gap-2">
                                <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                                </svg>
                                {error}
                            </p>
                        </div>
                    ) : (
                        <pre className="font-mono text-sm bg-gray-50 p-6 rounded-xl overflow-x-auto
                            max-h-[80vh] border-2 border-gray-200 shadow-inner whitespace-pre-wrap">
                            {currentLog || 'Загрузка...'}
                        </pre>
                    )}
                </div>
            </div>
        </div>
    )
}