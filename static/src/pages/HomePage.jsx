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

    const initializeFilters = (filtersData) => {
        const initialFilters = {}

        // Инициализация для multi-select
        filtersData['multi-select']?.forEach(filter => {
            initialFilters[filter.name] = []
        })

        // Инициализация для one-select
        filtersData['one-select']?.forEach(filter => {
            initialFilters[filter.name] = ''
        })

        // Инициализация для range
        filtersData.range?.forEach(filter => {
            initialFilters[filter.name] = {
                min: filter.from_value,
                max: filter.to_value
            }
        })

        // Инициализация для bool
        filtersData['bool-select']?.forEach(filter => {
            initialFilters[filter.name] = false
        })

        setSelectedFilters(initialFilters)
    }

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
                    if (isDefaultValue(currentValue, defaultValue, filterType)) return;

                    // Формируем объект фильтра
                    const filterObject = {
                        name: filterName,
                        value: formatFilterValue(currentValue, filterType)
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

    const formatFilterValue = (value, type) => {
        if (type === 'range') {
            return {
                from_value: Number(value.min) || 0,
                to_value: Number(value.max) || 0
            };
        }
        return value;
    };


    // Рендер разных типов фильтров
    const renderFilter = (filter, type) => {
        switch (type) {
            case 'range':
                return (
                    <div key={filter.name} className="mb-4">
                        <label className="block font-medium mb-2">{filter.name}</label>
                        <div className="flex gap-2">
                            <input
                                type="number"
                                placeholder="От"
                                value={selectedFilters[filter.name]?.min || ''}
                                onChange={(e) => setSelectedFilters(prev => ({
                                    ...prev,
                                    [filter.name]: {
                                        ...prev[filter.name],
                                        min: e.target.value
                                    }
                                }))}
                                className="w-1/2 p-2 border rounded"
                            />
                            <input
                                type="number"
                                placeholder="До"
                                value={selectedFilters[filter.name]?.max || ''}
                                onChange={(e) => setSelectedFilters(prev => ({
                                    ...prev,
                                    [filter.name]: {
                                        ...prev[filter.name],
                                        max: e.target.value
                                    }
                                }))}
                                className="w-1/2 p-2 border rounded"
                            />
                        </div>
                    </div>
                )

            case 'multi-select':
                return (
                    <div key={filter.name} className="mb-4">
                        <label className="block font-medium mb-2">{filter.name}</label>
                        <div className="grid grid-cols-2 gap-2">
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
                <div className="lg:col-span-1 bg-white p-6 rounded-xl shadow-lg">
                    <h2 className="text-xl font-bold mb-4">Фильтры</h2>

                    <div className="mb-4">
                        <label className="block font-medium mb-2">Категория</label>
                        <select
                            value={selectedCategory}
                            onChange={(e) => setSelectedCategory(e.target.value)}
                            className="w-full p-2 border rounded"
                        >
                            <option value="">Выберите категорию</option>
                            {categories.map(category => (
                                <option key={category} value={category}>{category}</option>
                            ))}
                        </select>
                    </div>

                    {filters?.range?.map(filter => renderFilter(filter, 'range'))}
                    {filters?.['multi-select']?.map(filter => renderFilter(filter, 'multi-select'))}
                    {filters?.['one-select']?.map(filter => renderFilter(filter, 'one-select'))}
                    {filters?.['bool-select']?.map(filter => renderFilter(filter, 'bool-select'))}

                    <div className="mb-4">
                        <label className="block font-medium mb-2">Сортировка</label>
                        <select
                            value={sortField}
                            onChange={(e) => setSortField(e.target.value)}
                            className="w-full p-2 border rounded mb-2"
                        >
                            <option value="">Без сортировки</option>
                            <option value="price">Цена</option>
                            <option value="rating">Рейтинг</option>
                        </select>
                        <select
                            value={sortOrder}
                            onChange={(e) => setSortOrder(e.target.value)}
                            className="w-full p-2 border rounded"
                        >
                            <option value="asc">По возрастанию</option>
                            <option value="desc">По убыванию</option>
                        </select>
                    </div>
                </div>

                {/* Основной контент */}
                <div className="lg:col-span-3">
                    <form onSubmit={handleSearch} className="mb-8">
                        <div className="flex gap-2">
                            <input
                                type="text"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                placeholder="Поиск товаров..."
                                className="flex-1 px-4 py-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
                            />
                            <button
                                type="submit"
                                className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-lg"
                            >
                                Найти
                            </button>
                        </div>
                    </form>

                    {/* Результаты поиска */}
                    {results && (
                        <div className="bg-white rounded-xl shadow-lg p-6">
                            <h2 className="text-2xl font-semibold mb-4">
                                Найдено записей: {results.length}
                            </h2>
                            <div className="space-y-4">
                                {results.length === 0 ? (
                                    <div className="p-4 text-center text-gray-500">
                                        Ничего не найдено
                                    </div>
                                ) : (
                                    results.map(({id, fields = {}, score}) => (
                                        <div key={id}
                                             className="p-4 border rounded-lg hover:bg-gray-50 transition-colors">
                                            <div className="flex flex-col gap-3">
                                                {/* Все поля объекта fields */}
                                                {Object.entries(fields).map(([key, value]) => (
                                                    <div key={key} className="flex gap-2">
                  <span className="text-gray-500 font-medium min-w-[120px]">
                    {formatKey(key)}:
                  </span>
                                                        <span className="text-gray-800 break-words flex-1">
                    {formatValue(value)}
                  </span>
                                                    </div>
                                                ))}

                                                {/* Дополнительная информация */}
                                                <div className="mt-2 pt-2 border-t">
                                                    <div className="flex justify-between items-center">
                                                        <div className="mt-2 text-xs text-gray-400">
                                                            <div>ID: {id}</div>
                                                            <button
                                                                onClick={() => openEditModal({id, fields})}
                                                                className="mt-1 text-blue-600 hover:text-blue-800 text-xs"
                                                            >
                                                                Редактировать →
                                                            </button>
                                                        </div>
                                                        {score && (
                                                            <span
                                                                className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
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

            {/*редактирование документа*/}
            {editingDoc && (
                <div
                    className="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center p-4 z-[100]"
                    onClick={(e) => e.target === e.currentTarget && setEditingDoc(null)}
                >
                    <div
                        className="bg-white rounded-xl p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto shadow-xl
                 animate-scale-in"
                    >
                        <h3 className="text-xl font-bold mb-4">Редактирование документа</h3>

                        {/* Динамическая форма */}
                        <div className="grid grid-cols-2 gap-4">
                            {Object.entries(formData).map(([key, value]) => (
                                <div key={key} className="space-y-2">
                                    <label className="block text-sm font-medium text-gray-700">
                                        {formatKey(key)}:
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
                                            className="w-full p-2 border rounded"
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
                                            className="w-full p-2 border rounded"
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
                                            className="w-full p-2 border rounded"
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
                                            className="w-full p-2 border rounded"
                                        />
                                    )}
                                </div>
                            ))}
                        </div>

                        {/* Кнопки */}
                        <div className="mt-6 flex gap-3 justify-end">
                            <button
                                onClick={() => setEditingDoc(null)}
                                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded transition-colors"
                            >
                                Отмена
                            </button>
                            <button
                                onClick={handleUpdate}
                                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
                            >
                                Сохранить
                            </button>
                        </div>
                    </div>
                </div>
            )}

        </div>
    )
}