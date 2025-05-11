import '../services/api'
import '../services/api'
import '../services/api'
import '../services/api'
import api from "../services/api.js";
import { useState, useEffect } from 'react'
import { useAppContext } from '../contexts/AppContext'

export default function Page2() {
    const {startLoading, stopLoading} = useAppContext()
    const [isModalOpen, setIsModalOpen] = useState(false)
    const [configType, setConfigType] = useState(null)
    const [configData, setConfigData] = useState('')
    const [error, setError] = useState(null)
    const [successMessage, setSuccessMessage] = useState(null)
    const [indexStatus, setIndexStatus] = useState({ isBuilt: false, isLoading: true })

    // Проверка статуса индекса
    const checkBuildStatus = async () => {
        try {
            const response = await api.checkBuildIndex()
            setIndexStatus({ isBuilt: response.data.isBuilded, isLoading: false })
        } catch (error) {
            setError('Ошибка: ' + error.message)
            setIndexStatus(prev => ({ ...prev, isLoading: false }))
        }
    }

    // Загрузка статуса при монтировании
    useEffect(() => {
        checkBuildStatus()
    }, [])

    // Пересборка индекса
    const handleRebuildIndex = async () => {
        try {
            startLoading()
            setError(null)
            await api.rebuildIndex()
            await checkBuildStatus()
            setSuccessMessage('Индекс успешно перестроен!')
        } catch (error) {
            setError('Ошибка пересборки индекса:' + error.message)
        } finally {
            stopLoading()
        }
    }

    // Откат конфига
    const handleRevertConfig = async () => {
        try {
            startLoading()
            setError(null)
            const response = await api.revertIndexConfig()
            setConfigData(JSON.stringify(response.data, null, 2))
            await checkBuildStatus()
            setSuccessMessage('Конфиг успешно откачен!')
        } catch (error) {
            setError('Ошибка отката конфигурации:' + error.message)
        } finally {
            stopLoading()
        }
    }

    // Загрузка конфига при открытии модалки
    useEffect(() => {
        const loadConfig = async () => {
            if (!isModalOpen || !configType) return

            try {
                startLoading()
                setError(null)
                setSuccessMessage(null)

                const response = await api.getConfig(configType)
                setConfigData(JSON.stringify(response.data, null, 2))
            } catch (error) {
                setError(error.response?.data || 'Ошибка загрузки конфигурации')
                setConfigData('')
            } finally {
                stopLoading()
            }
        }

        loadConfig()
    }, [isModalOpen, configType])

    // Обработчик сохранения конфига
    const handleSaveConfig = async () => {
        try {
            startLoading()
            setError(null)

            const parsedData = JSON.parse(configData)
            await api.updateConfig(configType, parsedData)

            setSuccessMessage('Конфигурация успешно сохранена!')
            setTimeout(() => setIsModalOpen(false), 1500)
        } catch (error) {
            setError(error.response?.data || 'Ошибка сохранения конфигурации')
        } finally {
            stopLoading()
        }
    }

    // Открытие модалки для конкретного типа конфига
    const openConfigModal = (type) => {
        setConfigType(type)
        setIsModalOpen(true)
    }

    return (
        <div className="max-w-md p-8 mx-auto">
            <h1 className="text-4xl font-bold mb-12 text-gray-800 text-center">
                Управление конфигурациями
                <span className="block mt-2 text-2xl text-gray-500">Выберите тип настройки</span>
            </h1>

            <div className="space-y-6">
                {/* Блок управления индексом */}
                <div className="bg-white p-6 rounded-xl shadow-lg border border-gray-100">
                    <div className="flex items-center gap-4 mb-4">
                        <button
                            onClick={() => openConfigModal('index')}
                            className="flex-1 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white
                                font-semibold py-4 px-8 rounded-xl transition-all duration-300 transform hover:scale-[1.02]
                                shadow-lg hover:shadow-xl active:scale-95 flex items-center justify-center gap-2"
                        >
                            <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                                <path d="M3 3h8v8H3V3zm0 10h8v8H3v-8zM13 3h8v8h-8V3zm0 10h8v8h-8v-8z"/>
                            </svg>
                            Настройки индекса
                        </button>

                        {/* Статус индекса */}
                        <div className="flex items-center gap-2">
                            {indexStatus.isLoading ? (
                                <div className="h-3 w-3 bg-gray-300 rounded-full animate-pulse"/>
                            ) : indexStatus.isBuilt ? (
                                <div className="h-3 w-3 bg-green-500 rounded-full shadow-green"/>
                            ) : (
                                <div className="h-3 w-3 bg-red-500 rounded-full animate-pulse"/>
                            )}
                        </div>
                    </div>

                    {/* Кнопки управления */}
                    <div className="grid grid-cols-2 gap-3">
                        <button
                            onClick={handleRebuildIndex}
                            className="bg-gradient-to-r from-purple-500 to-purple-600 hover:from-purple-600 hover:to-purple-700
                                text-white py-2 px-4 rounded-lg text-sm font-medium transition-all"
                            disabled={indexStatus.isLoading}
                        >
                            Пересобрать индекс
                        </button>
                        <button
                            onClick={handleRevertConfig}
                            className="bg-gradient-to-r from-gray-500 to-gray-600 hover:from-gray-600 hover:to-gray-700
                                text-white py-2 px-4 rounded-lg text-sm font-medium transition-all"
                            disabled={indexStatus.isLoading}
                        >
                            Откатить конфиг
                        </button>
                    </div>
                </div>

                <button
                    onClick={() => openConfigModal('filter')}
                    className="w-full bg-gradient-to-r from-green-500 to-green-600 hover:from-green-600 hover:to-green-700 text-white
                    font-semibold py-4 px-8 rounded-xl transition-all duration-300 transform hover:scale-[1.02]
                    shadow-lg hover:shadow-xl active:scale-95 flex items-center justify-center gap-2"
                >
                    <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                        <path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"/>
                    </svg>
                    Фильтры
                </button>

                <button
                    onClick={() => openConfigModal('ranking')}
                    className="w-full bg-gradient-to-r from-purple-500 to-purple-600 hover:from-purple-600 hover:to-purple-700 text-white
                    font-semibold py-4 px-8 rounded-xl transition-all duration-300 transform hover:scale-[1.02]
                    shadow-lg hover:shadow-xl active:scale-95 flex items-center justify-center gap-2"
                >
                    <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                        <path d="M12 2L4.5 20.29l.71.71L12 18l6.79 3 .71-.71L12 2z"/>
                    </svg>
                    Ранжирование
                </button>
            </div>

            {/* Модальное окно */}
            {isModalOpen && (
                <div
                    className="fixed inset-0 bg-black/20 backdrop-blur-md flex items-center justify-center p-4
                    animate-fade-in"
                    onClick={(e) => {
                        if (e.target === e.currentTarget) {
                            setIsModalOpen(false)
                            setError(null)
                            setSuccessMessage(null)
                        }
                    }}
                >
                    <div
                        className="bg-white rounded-2xl shadow-2xl w-full max-w-4xl max-h-[90vh] flex flex-col border
                        border-gray-100 animate-scale-in"
                        onClick={(e) => e.stopPropagation()}
                    >
                        <div className="p-8 flex-1 overflow-y-auto">
                            <div className="flex items-center justify-between mb-6">
                                <h3 className="text-2xl font-bold text-gray-800 capitalize">
                                <span
                                    className="bg-gradient-to-r from-blue-500 to-purple-600 bg-clip-text text-transparent">
                                    {configType}
                                </span>
                                </h3>
                                <button
                                    onClick={() => setIsModalOpen(false)}
                                    className="text-gray-400 hover:text-gray-600 transition-colors"
                                >
                                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                              d="M6 18L18 6M6 6l12 12"/>
                                    </svg>
                                </button>
                            </div>

                            <textarea
                                value={configData}
                                onChange={(e) => setConfigData(e.target.value)}
                                className="w-full h-[600px] px-6 py-4 border-2 rounded-xl font-mono text-sm
                                focus:ring-4 focus:ring-blue-200 focus:border-blue-500 resize-none
                                transition-all duration-200 border-gray-200"
                                placeholder="Загрузка конфигурации..."
                            />


                        </div>

                        <div className="p-6 bg-gray-50 flex justify-end gap-4 border-t border-gray-100">
                            {error && (
                                <div className="px-8 py-3 bg-red-50 rounded-xl border border-red-100 animate-fade-in">
                                    <p className="text-red-600 font-medium flex items-center gap-2">
                                        <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                                            <path
                                                d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                                        </svg>
                                        {error}
                                    </p>
                                </div>
                            )}

                            {successMessage && (
                                <div
                                    className="px-8 py-3 bg-green-50 rounded-xl border border-green-100 animate-fade-in">
                                    <p className="text-green-600 font-medium flex items-center gap-2">
                                        <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                                            <path
                                                d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                                        </svg>
                                        {successMessage}
                                    </p>
                                </div>
                            )}
                            <button
                                onClick={() => setIsModalOpen(false)}
                                className="px-8 py-3 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-xl
                                transition-colors duration-200 font-medium"
                            >
                                Отменить
                            </button>
                            <button
                                onClick={handleSaveConfig}
                                className="bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700
                                text-white px-8 py-3 rounded-xl font-medium transition-all duration-300
                                shadow-md hover:shadow-lg"
                            >
                                Сохранить изменения
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )

}