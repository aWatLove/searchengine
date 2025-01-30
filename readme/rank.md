# Ранжирование

## Конфигурация

- `field` - поле к которому применяется ранжирование
- `weight` - вес
- `boost_type` - преднастроенные формулы ранжирования
  - "catboostV2"
  - "logarithmic"
  - "custom'


Пример:
```json
{
  "boosts": [
    {
      "field": "brand",
      "weight": 5,
      "boost_type": "catboostV2"
    },
    {
      "field": "title",
      "weight": 3,
      "boost_type": "logarithmic"
    },
    {
      "field": "seller",
      "weight": 2,
      "boost_type": "catboost"
    },
    {
      "field": "color",
      "weight": 5,
      "boost_type": "custom",
      "formula": "$F^$W"
    }
  ]
}
```



### Custom формулы ранжирования

// format: 
`$F` - указывается field, 
`$W` - указывается weight. 

пример: `"$F^$W"`