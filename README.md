# Currency-Exchange-API

The API has 3 endpoints:

- http://localhost:8080/exchange/v1/exchangehistory/
- http://localhost:8080/exchange/v1/exchangeborder/
- http://localhost:8080/exchange/v1/diag/

If there is an error or something is wrong the API will respond with that information, an example of this can be that the external API is down, or the url is wrong, like the date or a country name that does not exist. For an example the restcountries.eu API has a problem when comparing two countries with same currency, this has been taken care of if the currency is EUR as the base in exchange/v1/exchangeborder/ is EUR, it will then return the exchange rate '1'.

If nothing fails the API will return the correct data in JSON.

# http://localhost:8080/exchange/v1/exchangehistory/

Type: Get

Format: 
http://localhost:8080/exchange/v1/exchangehistory/{:country_name}/{:begin_date-end_date}

Example request: http://localhost:8080/exchange/v1/exchangehistory/norway/2020-12-01-2021-01-31

Example response in JSON:
`
{
    "rates": {
        "2020-12-01": {
            "NOK": 1.1969
        },
        "2020-12-02": {
            "NOK": 1.1633
        },
        "2021-01-31": {
            "NOK": 1.1754
        }
    },
    "start_at": "2020-12-01",
    "base": "EUR",
    "end_at": "2021-01-31"
}
`

# http://localhost:8080/exchange/v1/exchangehistory/

Type: Get

Format: 
http://localhost:8080/exchange/v1/exchangeborder/{:country_name}{?limit={:number}}

Example request: http://localhost:8080/exchange/v1/exchangeborder/norway?limit=5

limit is optional

Example response in JSON:

`
{
    "rates": {
        "Sweden": {
            "currency": "SEK", 
            "rate": 1.1703
        },
        "Russia": {
            "currency": "RUB",
            "rate": 72.05
        }, 
    },
    "base": "NOK"
}
`

# http://localhost:8080/exchange/v1/diag/

Type: Get

Format: 
http://localhost:8080/exchange/v1/diag/

Example response in JSON:

`
{
   "exchangeratesapi": "200",
   "restcountries": "200",
   "version": "v1",
   "uptime": 57.1786634
}
`
