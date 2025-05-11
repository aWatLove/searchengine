import { BrowserRouter as Router, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import HomePage from './pages/HomePage'
import Page1 from './pages/Page1'
import Page2 from './pages/Page2'
import PageLogs from "./pages/PageLogs.jsx";
import PageMetrics from "./pages/PageMetrics.jsx";

export default function App() {
    return (
        <Router>
            <Routes>
                <Route path="/" element={<Layout />}>
                    <Route index element={<HomePage />} />
                    <Route path="index" element={<Page1 />} />
                    <Route path="config" element={<Page2 />} />
                    <Route path="logs" element={<PageLogs />} />
                    <Route path="metrics" element={<PageMetrics />} />
                </Route>
            </Routes>
        </Router>
    )
}