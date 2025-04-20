// import axios from 'axios'
//
// const api = axios.create({
//     baseURL: import.meta.env.VITE_API_URL || 'http://localhost:3001/api',
//     timeout: 5000,
// })
//
// export default {
//     // Пример метода
//     getData: () => api.get('/data'),
//     postData: (payload) => api.post('/data', payload),
// }

import axios from 'axios'

const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || 'http://localhost:3001/api',
    timeout: 5000,
})

export default {
    getAllData: () => api.get('/api/v1/getAllDoc'),
    postData: (payload) => api.post('/api/v1/addDoc', payload),
    getCategories: () => api.get('/api/v1/category')
        .then(res => res.data)
        .catch(() => []), // Возвращаем пустой массив при ошибке
    getFiltersByCategory: (category) =>
        api.get(`/api/v1/filtersByCategory?category=${encodeURIComponent(category)}`),
    search: (params) =>
        api.get('/api/v1/search', {
            params: {
                query: params.query,
                filters: JSON.stringify(params.filters),
                sortField: params.sortField,
                sortOrder: params.sortOrder
            }
        }),
    updateDoc: (docId, data) =>
        api.post(`/api/v1/updateDoc?docId=${docId}`, data)
}