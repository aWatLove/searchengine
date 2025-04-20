import { useState } from 'react'
import '../services/api'
import '../services/api'
import '../services/api'
import '../services/api'
import api from "../services/api.js";

export default function Page2() {
    const [formData, setFormData] = useState({ name: '', email: '' })
    const [response, setResponse] = useState(null)

    const handleSubmit = async (e) => {
        e.preventDefault()
        try {
            const result = await api.postData(formData)
            setResponse(result.data)
        } catch (error) {
            console.error('Ошибка отправки:', error)
        }
    }

    return (
        <div className="max-w-md mx-auto">
            <h2 className="text-2xl mb-6">Форма отправки данных</h2>

            <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                    <label className="block mb-1">Имя:</label>
                    <input
                        type="text"
                        value={formData.name}
                        onChange={(e) => setFormData({...formData, name: e.target.value})}
                        className="w-full p-2 border rounded"
                        required
                    />
                </div>

                <div>
                    <label className="block mb-1">Email:</label>
                    <input
                        type="email"
                        value={formData.email}
                        onChange={(e) => setFormData({...formData, email: e.target.value})}
                        className="w-full p-2 border rounded"
                        required
                    />
                </div>

                <button
                    type="submit"
                    className="w-full bg-green-500 hover:bg-green-600 text-white py-2 px-4 rounded"
                >
                    Отправить
                </button>
            </form>

            {response && (
                <div className="mt-6 p-4 bg-gray-100 rounded">
                    <h3 className="text-lg font-semibold mb-2">Ответ сервера:</h3>
                    <pre>{JSON.stringify(response, null, 2)}</pre>
                </div>
            )}
        </div>
    )
}