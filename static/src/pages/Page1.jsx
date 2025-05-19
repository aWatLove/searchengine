import {useEffect, useState} from 'react'
import {useAppContext} from '../contexts/AppContext'
import api from "../services/api.js";

const JsonRenderer = ({data}) => {
    console.log('Рендеринг данных:', data) // Добавлено логирование

    if (!data) return null

    if (Array.isArray(data)) {
        return (
            <div className="ml-4 border-l-2 border-gray-200 pl-4">
                <span className="text-gray-500">[</span>
                {data.map((item, index) => (
                    <div key={index} className="ml-4">
                        <JsonRenderer data={item}/>
                        {index < data.length - 1 && ','}
                    </div>
                ))}
                <span className="text-gray-500">]</span>
            </div>
        )
    }

    if (typeof data === 'object') {
        return (
            <div className="ml-4 border-l-2 border-gray-200 pl-4">
                <span className="text-gray-500">{'{'}</span>
                {Object.entries(data).map(([key, value], index, arr) => (
                    <div key={key} className="ml-4">
                        <span className="text-purple-600 font-medium">{key}:</span>
                        <JsonRenderer data={value}/>
                        {index < arr.length - 1 && ','}
                    </div>
                ))}
                <span className="text-gray-500">{'}'}</span>
            </div>
        )
    }

    return (
        <span className="ml-2">
      {typeof data === 'string' ? (
          <span className="text-green-600">"{data}"</span>
      ) : (
          <span className="text-blue-600">{data?.toString()}</span>
      )}
    </span>
    )
}

export default function Page1() {
    const {startLoading, stopLoading} = useAppContext()
    const [postData, setPostData] = useState('some struct')
    const [postResult, setPostResult] = useState(null)
    const [postError, setPostError] = useState(null)
    const [indexStruct, setIndexStruct] = useState([])
    const [docId, setDocId] = useState('')
    const [editData, setEditData] = useState('')
    const [editResult, setEditResult] = useState(null)
    const [editError, setEditError] = useState(null)
    const [deleteDocId, setDeleteDocId] = useState('')
    const [deleteResult, setDeleteResult] = useState(null)
    const [deleteError, setDeleteError] = useState(null)
    const [showConfirmModal, setShowConfirmModal] = useState(false)

    // Обработчик удаления документа
    const handleDeleteDoc = async (e) => {
        e.preventDefault()
        try {
            startLoading()
            setDeleteResult(null)
            setDeleteError(null)
            setShowConfirmModal(false)

            await api.deleteDoc(deleteDocId)

            setDeleteResult({
                message: 'Документ успешно удален!',
                id: deleteDocId
            })
            setDeleteDocId('') // Очищаем поле после удаления
        } catch (error) {
            setDeleteError(error.response?.data || error.message)
        } finally {
            stopLoading()
        }
    }
    // Новый обработчик для поиска документа
    const handleFindDoc = async () => {
        try {
            startLoading()
            setEditError(null)
            setEditResult(null)
            const response = await api.getDocumentById(docId)
            setEditData(JSON.stringify(response.data, null, 2))
            setEditError(null)
        } catch (error) {
            console.error('Ошибка при поиске документа:', error)
            setEditError(error.response?.data || 'Документ не найден')
            setEditData('')
        } finally {
            stopLoading()
        }
    }

    // Обработчик для обновления документа
    const handleUpdateDoc = async (e) => {
        e.preventDefault()

        try {
            startLoading()
            setEditError(null)
            setEditResult(null)

            const parsedData = JSON.parse(editData)
            const response = await api.updateDoc(docId, parsedData)

            setEditResult({
                message: 'Документ успешно обновлен!',
                data: response.data
            })
        } catch (error) {
            setEditError(error.response?.data || error.message)
        } finally {
            stopLoading()
        }
    }

    useEffect(() => {
        const loadIndexStruct = async () => {
            try {
                startLoading()
                const response = await api.getIndexStruct()

                // Проверка и нормализация данных
                const data = Array.isArray(response)
                    ? response
                    : response?.data || []

                setIndexStruct(data)
                setPostData(JSON.stringify(data, null, 2))
            } catch (error) {
                console.error('Failed to load categories:', error)
                setIndexStruct([])
            } finally {
                stopLoading()
            }
        }
        loadIndexStruct()
    }, [])


    const handleAddDoc = async (e) => {
        e.preventDefault()

        // Проверка на идентичность данных
        console.log('postData: ',postData)
        console.log('indexStruct: ', JSON.stringify(indexStruct, null, 2))
        if (postData === JSON.stringify(indexStruct, null, 2)) {
            const message = 'Пустой документ'

            setPostError(message)
            return
        }

        try {
            startLoading()
            setPostError(null)
            setPostResult(null)


            // Парсим JSON
            const parsedData = JSON.parse(postData)

            // Отправляем запрос
            const response = await api.postData(parsedData)

            setPostResult({
                id: response.data,
                message: 'Документ успешно добавлен!'
            })
        } catch (error) {
            // const message = error.response?.data?.message
            //     || error.message
            //     || 'Ошибка при отправке данных'

            const message = error.response.data

            setPostError(message)
        } finally {
            stopLoading()
        }
    }

    return (
        <div className="p-8">
            <div className="max-w-4xl mx-auto space-y-8">

                {/* Добавить документ */}
                <div className="bg-white rounded-2xl shadow-xl p-8 border border-gray-100">
                    <h2 className="text-3xl font-bold mb-6 bg-gradient-to-r from-green-600 to-blue-600 bg-clip-text text-transparent">
                        Добавить новый документ
                    </h2>

                    <form onSubmit={handleAddDoc} className="space-y-6">
                        <div>
                            <label className="block text-lg font-medium text-gray-800 mb-3 flex items-center gap-2">
                                <svg className="w-6 h-6 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                </svg>
                                JSON данные документа
                            </label>
                            <textarea
                                value={postData}
                                onChange={(e) => setPostData(e.target.value)}
                                className="w-full h-96 px-5 py-4 border-2 rounded-xl font-mono text-sm
                                focus:ring-4 focus:ring-blue-200 focus:border-blue-500 resize-none
                                transition-all duration-200 border-gray-200 shadow-sm"
                                placeholder={JSON.stringify(indexStruct, null, 2)}
                            />
                        </div>

                        <button
                            type="submit"
                            className="bg-gradient-to-r from-green-500 to-green-600 hover:from-green-600 hover:to-green-700
                            text-white font-semibold py-3 px-8 rounded-xl transition-all duration-300
                            transform hover:scale-[1.02] shadow-lg hover:shadow-xl active:scale-95"
                        >
                            Создать документ
                        </button>
                    </form>

                    {/* Результаты */}
                    {postResult && (
                        <div className="mt-6 p-5 bg-green-50 rounded-xl border-2 border-green-100 animate-fade-in">
                            <div className="flex items-center gap-3">
                                <div className="bg-green-100 p-2 rounded-full">
                                    <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                    </svg>
                                </div>
                                <div>
                                    <p className="text-green-600 font-semibold">{postResult.message}</p>
                                    <p className="text-green-700 mt-1">
                                        ID: <span className="font-mono bg-green-100 px-2 py-1 rounded">{postResult.id}</span>
                                    </p>
                                </div>
                            </div>
                        </div>
                    )}

                    {postError && (
                        <div className="mt-6 p-5 bg-red-50 rounded-xl border-2 border-red-100 animate-fade-in">
                            <div className="flex items-center gap-3">
                                <div className="bg-red-100 p-2 rounded-full">
                                    <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                    </svg>
                                </div>
                                <p className="text-red-600 font-semibold">Ошибка: {postError}</p>
                            </div>
                        </div>
                    )}
                </div>

                {/* Редактирование документа */}
                <div className="bg-white rounded-2xl shadow-xl p-8 border border-gray-100">
                    <h2 className="text-3xl font-bold mb-6 bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                        Редактировать документ
                    </h2>

                    <div className="space-y-6">
                        <div className="flex gap-4">
                            <input
                                type="text"
                                value={docId}
                                onChange={(e) => setDocId(e.target.value)}
                                placeholder="Введите ID документа"
                                className="flex-1 px-5 py-3 border-2 rounded-xl focus:ring-4
                                focus:ring-blue-200 focus:border-blue-500 border-gray-200
                                shadow-sm transition-all duration-200"
                            />
                            <button
                                onClick={handleFindDoc}
                                className="bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700
                                text-white font-semibold py-3 px-8 rounded-xl transition-all duration-300
                                transform hover:scale-[1.02] shadow-lg hover:shadow-xl active:scale-95"
                            >
                                Найти документ
                            </button>
                        </div>

                        {editData && (
                            <form onSubmit={handleUpdateDoc} className="space-y-6">
                            <textarea
                                value={editData}
                                onChange={(e) => setEditData(e.target.value)}
                                className="w-full h-96 px-5 py-4 border-2 rounded-xl font-mono text-sm
                                    focus:ring-4 focus:ring-blue-200 focus:border-blue-500 resize-none
                                    transition-all duration-200 border-gray-200 shadow-sm"
                            />
                                <button
                                    type="submit"
                                    className="bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700
                                    text-white font-semibold py-3 px-8 rounded-xl transition-all duration-300
                                    transform hover:scale-[1.02] shadow-lg hover:shadow-xl active:scale-95"
                                >
                                    Сохранить изменения
                                </button>
                            </form>
                        )}

                        {/* Результаты редактирования */}
                        {editResult && (
                            <div className="mt-6 p-5 bg-green-50 rounded-xl border-2 border-green-100 animate-fade-in">
                                <div className="flex items-center gap-3">
                                    <div className="bg-green-100 p-2 rounded-full">
                                        <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                        </svg>
                                    </div>
                                    <div>
                                        <p className="text-green-600 font-semibold">{editResult.message}</p>

                                    </div>
                                </div>
                            </div>
                        )}

                        {editError && (
                            <div className="mt-6 p-5 bg-red-50 rounded-xl border-2 border-red-100 animate-fade-in">
                                <div className="flex items-center gap-3">
                                    <div className="bg-red-100 p-2 rounded-full">
                                        <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                        </svg>
                                    </div>
                                    <p className="text-red-600 font-semibold">Ошибка: {editError}</p>
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                {/* Удаление документа */}
                <div className="bg-white rounded-2xl shadow-xl p-8 border border-gray-100">
                    <h2 className="text-3xl font-bold mb-6 bg-gradient-to-r from-red-600 to-pink-600 bg-clip-text text-transparent">
                        Удалить документ
                    </h2>

                    <form onSubmit={(e) => {
                        e.preventDefault()
                        if (deleteDocId.length !== 0) {
                            setShowConfirmModal(true)
                        }
                    }} className="space-y-6">
                        <div className="flex gap-4">
                            <input
                                type="text"
                                value={deleteDocId}
                                onChange={(e) => setDeleteDocId(e.target.value)}
                                placeholder="Введите ID документа"
                                className="flex-1 px-5 py-3 border-2 rounded-xl focus:ring-4 focus:ring-red-200 focus:border-red-500 border-gray-200 shadow-sm transition-all duration-200"
                            />
                            <button
                                type="submit"
                                className="bg-gradient-to-r from-red-500 to-red-600 hover:from-red-600 hover:to-red-700 text-white font-semibold py-3 px-8 rounded-xl transition-all duration-300 transform hover:scale-[1.02] shadow-lg hover:shadow-xl active:scale-95"
                            >
                                Удалить документ
                            </button>
                        </div>

                        {/* Результаты удаления */}
                        {deleteResult && (
                            <div className="mt-6 p-5 bg-green-50 rounded-xl border-2 border-green-100 animate-fade-in">
                                <div className="flex items-center gap-3">
                                    <div className="bg-green-100 p-2 rounded-full">
                                        <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                        </svg>
                                    </div>
                                    <div>
                                        <p className="text-green-600 font-semibold">{deleteResult.message}</p>
                                        <p className="text-green-700 mt-1">
                                            ID: <span className="font-mono bg-green-100 px-2 py-1 rounded">{deleteResult.id}</span>
                                        </p>
                                    </div>
                                </div>
                            </div>
                        )}

                        {deleteError && (
                            <div className="mt-6 p-5 bg-red-50 rounded-xl border-2 border-red-100 animate-fade-in">
                                <div className="flex items-center gap-3">
                                    <div className="bg-red-100 p-2 rounded-full">
                                        <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                        </svg>
                                    </div>
                                    <p className="text-red-600 font-semibold">Ошибка: {deleteError}</p>
                                </div>
                            </div>
                        )}
                    </form>

                    {showConfirmModal && (
                        <div className="fixed inset-0 bg-black/20 backdrop-blur-sm bg-opacity-50 flex items-center justify-center z-50">
                            <div className="bg-white rounded-2xl p-6 max-w-md w-full shadow-xl border">
                                <h3 className="text-xl font-bold text-gray-900 mb-4">Подтвердите удаление</h3>
                                <p className="text-gray-600 mb-6">Вы уверены, что хотите удалить документ с ID: <span className="font-mono">{deleteDocId}</span>?</p>
                                <div className="flex justify-end gap-4">
                                    <button
                                        className="px-5 py-2 rounded-xl border border-gray-300 hover:bg-gray-100 transition"
                                        onClick={() => setShowConfirmModal(false)}
                                    >
                                        Отмена
                                    </button>
                                    <button
                                        className="px-5 py-2 rounded-xl bg-red-600 text-white hover:bg-red-700 transition"
                                        onClick={handleDeleteDoc}
                                    >
                                        Удалить
                                    </button>
                                </div>
                            </div>
                        </div>
                    )}
                </div>



            </div>
        </div>
    )
}