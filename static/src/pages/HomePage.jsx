import {useState, useEffect} from 'react'
import {useAppContext} from '../contexts/AppContext'
import api from '../services/api'

export default function HomePage() {
    const {startLoading, stopLoading} = useAppContext()
    const [searchQuery, setSearchQuery] = useState('')
    const [categories, setCategories] = useState([])
    const [selectedCategory, setSelectedCategory] = useState('')
    const [filters, setFilters] = useState(null)
    const [selectedFilters, setSelectedFilters] = useState({})
    const [sortField, setSortField] = useState('')
    const [sortOrder, setSortOrder] = useState('asc')
    const [results, setResults] = useState(null)
    const [editingDoc, setEditingDoc] = useState(null)
    const [formData, setFormData] = useState({})


    // Загрузка категорий
    useEffect(() => {
        const loadCategories = async () => {
            try {
                startLoading()
                const response = await api.getCategories()

                // Проверка и нормализация данных
                const data = Array.isArray(response)
                    ? response
                    : response?.data || []

                setCategories(data)
            } catch (error) {
                console.error('Failed to load categories:', error)
                setCategories([])
            } finally {
                stopLoading()
            }
        }
        loadCategories()
    }, [])

    // Загрузка фильтров при выборе категории
    useEffect(() => {
        const loadFilters = async () => {
            if (!selectedCategory) return

            try {
                startLoading()
                const response = await api.getFiltersByCategory(selectedCategory)
                setFilters(response.data)
                initializeFilters(response.data)
            } catch (error) {
                console.error('Ошибка загрузки фильтров:', error)
            } finally {
                stopLoading()
            }
        }
        loadFilters()
    }, [selectedCategory])


// Обновленная функция инициализации фильтров
    const initializeFilters = (filtersData) => {
        const initialFilters = {};

        // Инициализация для multi-select
        filtersData['multi-select']?.forEach(filter => {
            initialFilters[filter.name] = [];
        });

        // Инициализация для one-select
        filtersData['one-select']?.forEach(filter => {
            initialFilters[filter.name] = '';
        });

        // Инициализация для range (сохраняем тип)
        filtersData.range?.forEach(filter => {
            initialFilters[filter.name] = {
                type: filter.type, // Сохраняем тип фильтра
                min: filter.from_value,
                max: filter.to_value
            };
        });

        // Инициализация для bool
        filtersData['bool-select']?.forEach(filter => {
            initialFilters[filter.name] = false;
        });

        setSelectedFilters(initialFilters);
    };

    const handleSearch = async (e) => {
        e.preventDefault();
        try {
            startLoading();

            const requestFilters = {};

            // Добавляем категорию, если выбрана
            if (selectedCategory) {
                requestFilters.category = selectedCategory;
            }

            // Группируем фильтры по типам
            const filterGroups = {
                'multi-select': [],
                'one-select': [],
                range: [],
                'bool-select': []
            };

            // Обрабатываем каждый фильтр
            Object.entries(filterGroups).forEach(([filterType]) => {
                filters?.[filterType]?.forEach(filterDef => {
                    const filterName = filterDef.name;
                    const currentValue = selectedFilters[filterName];
                    const defaultValue = getDefaultFilterValue(filterType, filterDef);

                    // Пропускаем фильтр, если значение дефолтное
                    if ((filterType !== 'range') && (isDefaultValue(currentValue, defaultValue, filterType))) return;

                    // Формируем объект фильтра с учетом типа
                    const filterObject = filterType === 'range' ? {
                        name: filterName,
                        type: filterDef.type, // Добавляем тип из описания фильтра
                        from_value: String(currentValue.min) || '0',
                        to_value: String(currentValue.max) || '0'
                    } : {
                        name: filterName,
                        value: currentValue
                    };

                    filterGroups[filterType].push(filterObject);
                });
            });

            // Добавляем только непустые группы
            Object.entries(filterGroups).forEach(([key, values]) => {
                if (values.length > 0) {
                    requestFilters[key] = values;
                }
            });

            // Формируем параметры запроса
            const params = {
                query: searchQuery,
                ...(Object.keys(requestFilters).length > 0 && {filters: requestFilters}),
                sortField,
                sortOrder
            };

            console.log('Final params:', JSON.stringify(params, null, 2));

            const response = await api.search(params);
            setResults(response.data);
        } catch (error) {
            console.error('Search error:', error.response?.data || error.message);
        } finally {
            stopLoading();
        }
    };

// Вспомогательные функции
    const getDefaultFilterValue = (type, filterDef) => {
        switch (type) {
            case 'multi-select':
                return [];
            case 'one-select':
                return '';
            case 'range':
                return {
                    min: filterDef.from_value,
                    max: filterDef.to_value
                };
            case 'bool-select':
                return false;
            default:
                return null;
        }
    };

    const isDefaultValue = (current, defaultVal, type) => {
        if (type === 'range') {
            return current.min === defaultVal.min && current.max === defaultVal.max;
        }
        return JSON.stringify(current) === JSON.stringify(defaultVal);
    };


    // Рендер разных типов фильтров
    const renderFilter = (filter, type) => {
        switch (type) {
            case 'range':
                return (
                    <div key={filter.name} className="mb-4">
                        <label className="block font-medium mb-2">
                            {filter.name}
                        </label>
                        <div className="flex flex-col gap-2">
                            <div className="flex items-center gap-2">
                                <span className="text-gray-600 w-7">От:</span>
                                <input
                                    type={filter.type === 'timestamp' ? 'date' : 'number'}
                                    value={selectedFilters[filter.name]?.min || ''}
                                    onChange={(e) => setSelectedFilters(prev => ({
                                        ...prev,
                                        [filter.name]: {
                                            ...prev[filter.name],
                                            min: e.target.value
                                        }
                                    }))}
                                    className="flex-1 p-2 w-1/3 border rounded"
                                />
                            </div>
                            <div className="flex items-center gap-2">
                                <span className="text-gray-600 w-7">До:</span>
                                <input
                                    type={filter.type === 'timestamp' ? 'date' : 'number'}
                                    value={selectedFilters[filter.name]?.max || ''}
                                    onChange={(e) => setSelectedFilters(prev => ({
                                        ...prev,
                                        [filter.name]: {
                                            ...prev[filter.name],
                                            max: e.target.value
                                        }
                                    }))}
                                    className="flex-1 p-2 w-1/3 border rounded"
                                />
                            </div>
                        </div>
                    </div>
                )
            case 'multi-select':
                return (
                    <div key={filter.name} className="mb-4">
                        <label className="block font-medium mb-2">{filter.name}</label>
                        <div className="grid grid-cols-1 gap-1">
                            {filter.value.map(option => (
                                <label key={option} className="flex items-center">
                                    <input
                                        type="checkbox"
                                        checked={selectedFilters[filter.name]?.includes(option)}
                                        onChange={(e) => {
                                            const newValues = e.target.checked
                                                ? [...selectedFilters[filter.name], option]
                                                : selectedFilters[filter.name].filter(v => v !== option)
                                            setSelectedFilters(prev => ({
                                                ...prev,
                                                [filter.name]: newValues
                                            }))
                                        }}
                                        className="mr-2"
                                    />
                                    {option}
                                </label>
                            ))}
                        </div>
                    </div>
                )

            case 'one-select':
                return (
                    <div key={filter.name} className="mb-4">
                        <label className="block font-medium mb-2">{filter.name}</label>
                        <select
                            value={selectedFilters[filter.name] || ''}
                            onChange={(e) => setSelectedFilters(prev => ({
                                ...prev,
                                [filter.name]: e.target.value
                            }))}
                            className="w-full p-2 border rounded"
                        >
                            <option value="">Не выбрано</option>
                            {filter.value.map(option => (
                                <option key={option} value={option}>{option}</option>
                            ))}
                        </select>
                    </div>
                )

            case 'bool-select':
                return (
                    <div key={filter.name} className="mb-4">
                        <label className="flex items-center">
                            <input
                                type="checkbox"
                                checked={selectedFilters[filter.name] || false}
                                onChange={(e) => setSelectedFilters(prev => ({
                                    ...prev,
                                    [filter.name]: e.target.checked
                                }))}
                                className="mr-2"
                            />
                            {filter.name}
                        </label>
                    </div>
                )
        }
    }

    // Форматирование ключей
    const formatKey = (key) => {
        // Примеры преобразований:
        // 'productName' → 'Product Name'
        // 'top-seller' → 'Top Seller'
        // 'created_at' → 'Created At'
        return key
            .replace(/([A-Z])/g, ' $1')
            .replace(/[-_]/g, ' ')
            .replace(/\b\w/g, l => l.toUpperCase())
            .trim()
    }

// Форматирование значений
    const formatValue = (value) => {
        if (value === null || value === undefined) return '-'

        if (Array.isArray(value)) {
            return value.join(', ') || 'Пусто'
        }

        if (typeof value === 'object') {
            return JSON.stringify(value, null, 2)
        }

        if (typeof value === 'boolean') {
            return value ? 'Да' : 'Нет'
        }

        if (typeof value === 'number') {
            return value.toLocaleString('ru-RU')
        }

        return value.toString()
    }

    //Функция открытия модального окна
    const openEditModal = (doc) => {
        setEditingDoc(doc)
        // Преобразуем категорию в массив, если это строка
        const initialData = {
            ...doc.fields,
            category: Array.isArray(doc.fields.category)
                ? doc.fields.category
                : [doc.fields.category].filter(Boolean)
        }
        setFormData(initialData)
    }

    //  Функция отправки изменений
    const handleUpdate = async () => {
        try {
            const response = await api.updateDoc(editingDoc.id, formData)
            setResults(results.map(item =>
                item.id === editingDoc.id ? {...item, fields: response.data} : item
            ))
            setEditingDoc(null)
        } catch (error) {
            console.error('Ошибка обновления:', error)
        }
    }

    return (
        <div className="p-8 max-w-7xl mx-auto">
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
                {/* Блок фильтров */}
                <div className="lg:col-span-1 bg-white p-6 rounded-2xl shadow-xl border border-gray-100">
                    <h2 className="text-2xl font-bold mb-6 bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                        Фильтры
                    </h2>

                    <div className="space-y-6">
                        <div className="space-y-3">
                            <label className="block text-sm font-medium text-gray-700 flex items-center gap-2">
                                <svg className="w-5 h-5 text-gray-500" fill="none" stroke="currentColor"
                                     viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                          d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"/>
                                </svg>
                                Категория
                            </label>
                            <select
                                value={selectedCategory}
                                onChange={(e) => setSelectedCategory(e.target.value)}
                                className="w-full px-4 py-3 border-2 rounded-xl focus:ring-4 focus:ring-purple-200 focus:border-purple-500 transition-all"
                            >
                                <option value="">Все категории</option>
                                {categories.map(category => (
                                    <option key={category} value={category} className="capitalize">{category}</option>
                                ))}
                            </select>
                        </div>

                        {filters?.range?.map(filter => renderFilter(filter, 'range'))}
                        {filters?.['multi-select']?.map(filter => renderFilter(filter, 'multi-select'))}
                        {filters?.['one-select']?.map(filter => renderFilter(filter, 'one-select'))}
                        {filters?.['bool-select']?.map(filter => renderFilter(filter, 'bool-select'))}

                        <div className="space-y-4 pt-4 border-t border-gray-100">
                            <label className="block text-sm font-medium text-gray-700 flex items-center gap-2">
                                <svg className="w-5 h-5 text-gray-500" fill="none" stroke="currentColor"
                                     viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                          d="M3 4h13M3 8h9m-9 4h9m5-4v12m0 0l-4-4m4 4l4-4"/>
                                </svg>
                                Сортировка
                            </label>
                            <div className="space-y-2">
                                <select
                                    value={sortField}
                                    onChange={(e) => setSortField(e.target.value)}
                                    className="w-full px-4 py-3 border-2 rounded-xl focus:ring-4 focus:ring-purple-200"
                                >
                                    <option value="">Без сортировки</option>
                                    <option value="price">Цена</option>
                                </select>
                                <select
                                    value={sortOrder}
                                    onChange={(e) => setSortOrder(e.target.value)}
                                    className="w-full px-4 py-3 border-2 rounded-xl focus:ring-4 focus:ring-purple-200"
                                >
                                    <option value="asc">По возрастанию</option>
                                    <option value="desc">По убыванию</option>
                                </select>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Основной контент */}
                <div className="lg:col-span-3">
                    <form onSubmit={handleSearch} className="mb-8">
                        <div className="flex gap-2 relative">
                            <input
                                type="text"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                placeholder="Поиск документов..."
                                className="flex-1 px-6 py-4 border-2 rounded-2xl focus:ring-4 focus:ring-blue-200
                                focus:border-blue-500 text-lg shadow-sm transition-all"
                            />
                            <button
                                type="submit"
                                className="bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700
                                text-white px-8 py-4 rounded-2xl font-semibold transition-all duration-300
                                transform hover:scale-[1.02] shadow-lg hover:shadow-xl active:scale-95"
                            >
                                Найти
                            </button>
                        </div>
                    </form>

                    {/* Результаты поиска */}
                    {results && (
                        <div className="bg-white rounded-2xl shadow-xl p-6 border border-gray-100">
                            <h2 className="text-2xl font-bold mb-6 bg-gradient-to-r from-gray-800 to-gray-600 bg-clip-text text-transparent">
                                Найдено документов: <span className="text-blue-600">{results.length}</span>
                            </h2>
                            <div className="space-y-4">
                                {results.length === 0 ? (
                                    <div className="p-8 text-center bg-gray-50 rounded-xl">
                                        <div className="text-gray-400 text-xl mb-2">😕</div>
                                        <p className="text-gray-500 font-medium">Ничего не найдено</p>
                                    </div>
                                ) : (
                                    results.map(({id, fields = {}, score}) => (
                                        <div
                                            key={id}
                                            className="group p-6 border-2 rounded-xl hover:border-blue-200
                                            hover:shadow-lg transition-all duration-300 cursor-pointer"
                                        >
                                            <div className="flex flex-col gap-3">
                                                {Object.entries(fields).map(([key, value]) => (
                                                    <div key={key} className="flex gap-3 items-start">
                                                    <span className="text-sm font-medium text-gray-500 min-w-[140px]
                                                        px-3 py-1 bg-gray-50 rounded-full">
                                                        {formatKey(key)}
                                                    </span>
                                                        <span className="text-gray-800 break-words flex-1">
                                                        {formatValue(value)}
                                                    </span>
                                                    </div>
                                                ))}

                                                <div className="mt-4 pt-4 border-t border-gray-100">
                                                    <div className="flex justify-between items-center">
                                                        <div className="flex items-center gap-3">
                                                            <span className="text-xs text-gray-400">ID: {id}</span>
                                                            <button
                                                                onClick={() => openEditModal({id, fields})}
                                                                className="text-blue-600 hover:text-blue-800 text-sm
                                                                flex items-center gap-1 transition-colors"
                                                            >
                                                                <svg className="w-4 h-4" fill="none"
                                                                     stroke="currentColor" viewBox="0 0 24 24">
                                                                    <path strokeLinecap="round" strokeLinejoin="round"
                                                                          strokeWidth={2}
                                                                          d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/>
                                                                </svg>
                                                                Редактировать
                                                            </button>
                                                        </div>
                                                        {score && (
                                                            <span className="text-xs bg-gradient-to-r from-blue-500 to-purple-500
                                                            text-white px-3 py-1 rounded-full shadow-sm">
                                                            Score: {Number(score).toFixed(2)}
                                                        </span>
                                                        )}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {/* Модальное окно редактирования */}
            {editingDoc && (
                <div
                    className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center p-4 z-50
                    animate-fade-in"
                    onClick={(e) => e.target === e.currentTarget && setEditingDoc(null)}
                >
                    <div
                        className="bg-white rounded-2xl p-8 max-w-2xl w-full max-h-[90vh] overflow-y-auto shadow-2xl
                        border border-gray-100 animate-scale-in"
                    >
                        <div className="flex justify-between items-center mb-6">
                            <h3 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600
                            bg-clip-text text-transparent">
                                Редактирование документа
                            </h3>
                            <button
                                onClick={() => setEditingDoc(null)}
                                className="text-gray-400 hover:text-gray-600 transition-colors"
                            >
                                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                          d="M6 18L18 6M6 6l12 12"/>
                                </svg>
                            </button>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {Object.entries(formData).map(([key, value]) => (
                                <div key={key} className="space-y-2">
                                    <label className="block text-sm font-medium text-gray-700">
                                        {formatKey(key)}
                                    </label>

                                    {key === 'category' ? (
                                        <input
                                            value={value.join(', ')}
                                            onChange={(e) =>
                                                setFormData(prev => ({
                                                    ...prev,
                                                    [key]: e.target.value.split(',').map(s => s.trim())
                                                }))
                                            }
                                            className="w-full px-4 py-3 border-2 rounded-xl focus:ring-4 focus:ring-blue-200"
                                        />
                                    ) : typeof value === 'boolean' ? (
                                        <select
                                            value={value}
                                            onChange={(e) =>
                                                setFormData(prev => ({
                                                    ...prev,
                                                    [key]: e.target.value === 'true'
                                                }))
                                            }
                                            className="w-full px-4 py-3 border-2 rounded-xl focus:ring-4 focus:ring-blue-200"
                                        >
                                            <option value="true">Да</option>
                                            <option value="false">Нет</option>
                                        </select>
                                    ) : Array.isArray(value) ? (
                                        <input
                                            value={value.join(', ')}
                                            onChange={(e) =>
                                                setFormData(prev => ({
                                                    ...prev,
                                                    [key]: e.target.value.split(',').map(s => s.trim())
                                                }))
                                            }
                                            className="w-full px-4 py-3 border-2 rounded-xl focus:ring-4 focus:ring-blue-200"
                                        />
                                    ) : (
                                        <input
                                            type={typeof value === 'number' ? 'number' : 'text'}
                                            value={value}
                                            onChange={(e) =>
                                                setFormData(prev => ({
                                                    ...prev,
                                                    [key]: typeof value === 'number'
                                                        ? Number(e.target.value)
                                                        : e.target.value
                                                }))
                                            }
                                            className="w-full px-4 py-3 border-2 rounded-xl focus:ring-4 focus:ring-blue-200"
                                        />
                                    )}
                                </div>
                            ))}
                        </div>

                        <div className="mt-8 flex gap-4 justify-end border-t border-gray-100 pt-6">
                            <button
                                onClick={() => setEditingDoc(null)}
                                className="px-6 py-3 text-gray-600 hover:bg-gray-50 rounded-xl font-medium
                                transition-colors"
                            >
                                Отмена
                            </button>
                            <button
                                onClick={handleUpdate}
                                className="bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700
                                text-white px-8 py-3 rounded-xl font-semibold transition-all shadow-md hover:shadow-lg"
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