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
    getIndexStruct: () => api.get('/api/v1/indexStruct'),
    deleteDoc: (docId) => api.delete(`/api/v1/deleteDoc?docId=${docId}`),
    getDocumentById: (docId) => api.get(`/api/v1/getDocId?docId=${docId}`),
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
        api.post(`/api/v1/updateDoc?docId=${docId}`, data),

    getConfig: (type) => api.get(`/api/v1/getConfig/${type}`),
    updateConfig: (type, data) => api.post(`/api/v1/config/${type}`, data),

    checkBuildIndex: () => api.get('/api/v1/config/index/isbuild'),
    rebuildIndex: () => api.get('/api/v1/rebuild'),
    revertIndexConfig: () => api.get('/api/v1/config/index/revert'),
}