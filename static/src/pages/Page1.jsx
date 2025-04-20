import {useState} from 'react'
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
    const [data, setData] = useState(null)
    const [postData, setPostData] = useState('')
    const [postResult, setPostResult] = useState(null)
    const [postError, setPostError] = useState(null)

    const handleGetData = async () => {
        try {
            startLoading()
            const response = await api.getAllData()
            console.log('Полные данные ответа:', response)
            setData(response.data)
        } catch (error) {
            console.error('Полная ошибка:', error)
        } finally {
            stopLoading()
        }
    }

    const handleAddDoc = async (e) => {
        e.preventDefault()
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



                {/* Получить все документы */}
                <div className="mt-6 p-6  min-h-[50vh] max-h-[85vh] bg-white rounded-xl shadow-lg">
                    <h2 className="text-2xl font-bold mb-4">Получить все документы</h2>

                    <button
                        onClick={handleGetData}
                        className="bg-green-600 hover:bg-green-700 text-white
                font-medium py-2 px-6 rounded-lg transition-colors mb-4"
                    >
                        Получить
                    </button>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                    Результат запроса:
                    </label>
                    <div className="min-h-[25vh] max-h-[60vh] overflow-y-auto ring-1 ring-gray-500 rounded-lg p-4">
                        {data && (
                            <div>

                                <JsonRenderer data={data}/>
                            </div>

                        )}
                    </div>
                </div>

                {/* Добавить документ */}
                <div className="bg-white rounded-xl shadow-lg p-6">
                    <h2 className="text-2xl font-bold mb-4">Добавить новый документ</h2>

                    <form onSubmit={handleAddDoc} className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Введите данные в JSON-формате:
                            </label>
                            <textarea
                                value={postData}
                                onChange={(e) => setPostData(e.target.value)}
                                className="w-full h-48 px-3 py-2 border rounded-lg
                  font-mono text-sm focus:ring-2 focus:ring-blue-500
                  focus:border-blue-500 resize-none"
                                placeholder={'{\n  "field": "value"\n}'}
                            />
                        </div>

                        <button
                            type="submit"
                            className="bg-green-600 hover:bg-green-700 text-white
                font-medium py-2 px-6 rounded-lg transition-colors"
                        >
                            Отправить данные
                        </button>
                    </form>

                    {/* Блок с результатом */}
                    {postResult && (
                        <div className="mt-6 p-4 bg-green-50 rounded-lg">
                            <p className="text-green-600 font-medium">
                                ✅ {postResult.message}
                            </p>
                            <p className="mt-2 text-green-800">
                                ID: <span className="font-mono">{postResult.id}</span>
                            </p>
                        </div>
                    )}

                    {postError && (
                        <div className="mt-6 p-4 bg-red-50 rounded-lg">
                            <p className="text-red-600 font-medium">
                                ❌ Ошибка: {postError}
                            </p>
                        </div>
                    )}
                </div>
            </div>
        </div>

    )
}