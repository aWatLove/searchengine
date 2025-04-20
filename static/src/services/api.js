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
    postData: (payload) => api.post('/api/v1/addDoc', payload)
}