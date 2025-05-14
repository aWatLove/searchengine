import { Outlet, Link } from 'react-router-dom'
import '../index.css'


export default function Layout() {
    return (
        <div className="flex min-h-screen">
            {/* Боковое меню */}
            <nav className="w-64 bg-gray-800 p-4 text-white">
                <div className="text-xl font-bold mb-6">Меню</div>
                <ul className="space-y-2">
                    <li>
                        <Link
                            to="/"
                            className="block px-4 py-2 hover:bg-gray-700 rounded transition-colors"
                        >
                            ПОИСК
                        </Link>
                    </li>
                    <li>
                        <Link
                            to="/index"
                            className="block px-4 py-2 hover:bg-gray-700 rounded transition-colors"
                        >
                            ИНДЕКС
                        </Link>
                    </li>
                    <li>
                        <Link
                            to="/config"
                            className="block px-4 py-2 hover:bg-gray-700 rounded transition-colors"
                        >
                            КОНФИГУРАЦИИ
                        </Link>
                    </li>
                    <li>
                        <Link
                            to="/logs"
                            className="block px-4 py-2 hover:bg-gray-700 rounded transition-colors"
                        >
                            ЛОГИ
                        </Link>
                    </li>
                    <li>
                        <Link
                            to="/metrics"
                            className="block px-4 py-2 hover:bg-gray-700 rounded transition-colors"
                        >
                            МЕТРИКИ
                        </Link>
                    </li>
                </ul>
            </nav>

            {/* Основной контент */}
            <main className="flex-1 p-8 bg-gray-100">
                <Outlet />
            </main>
        </div>
    )
}