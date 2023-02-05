# Fenix: A simple convenient messaging service

Fenix is an easy to use messaging service geared towards making instant messaging available to a wider audience of users.



# API Reference

## Authentication

Fenix currently uses token authentication.

There are two endpoints to authenticate with:`/login` and `/register`

### To Authenticate
1.  Send a POST request to an authentication endpoint, with your username and password JSON encoded in the body
2.  Fenix will respond with a JSON POST body with a token and user ID.
3.  Send a Websocket Upgrade request to `/upgrade?t=YOUR_TOKEN_HERE&id=YOUR_ID_HERE`
4.  Congrats!  You are securely connected to Fenix.

* * *

### `/login`

#### Description:

Connects a user to Fenix with an existing username and password.

#### Request:

Valid Basic Auth header

#### Responses

| **Scenario** | **Response** |
| --- | --- |
| Invalid Basic Auth header | 400 Bad Request |
| User doesn't exist / invalid password | 403 Forbidden |
| Error upgrading connection | 500 Internal Server Error |
| Successful login | Connection upgraded to websocket, listening for messages |

### `/register`

#### Description:

Connects a user to Fenix without an existing username and password.

#### Request:

Valid Basic Auth header

#### Responses

| **Scenario** | **Response** |
| --- | --- |
| Invalid Basic Auth header | 400 Bad Request |
| Username already taken | 409 Conflict |
| Error inserting into database | 500 Internal Server Error |
| Error upgrading connection | 500 Internal Server Error |
| Successful registration | Connection upgraded to websocket, listening for messages |


## Identification

### `whoami`

#### Description:

Returns the ID and username of the user.

#### Request:

``` json
{
    "type":   "whoami"
}

```

#### Response

``` json
{
    "type": "whoami",
    "id": "63c74c018cb827613b1e6bea",
    "nick": "piesquared"
}

```

* * *

## Messaging

### `msg_send`

#### Description:

Sends a chat message to Fenix

#### Request:

``` json
{
    "type": "msg_send",
    "msg": "Welcome to Fenix!"
}

```

#### Response

###### Successful message sent

``` json
{
    "type": "msg_broadcast",
    "m_id": "63c74f428cb827613b1e6beb",
    "author": {
        "ID": "63c74c018cb827613b1e6bea",
        "Username": "piesquared"
    },
    "msg": "Welcome to Fenix!",
    "time": 1674006338288360000
}

```

| **Scenario** | **Response** |
| --- | --- |
| Invalid JSON | JSONDecodeError |
| msg field in msg_send was empty | MessageEmpty |
| Error inserting message into database | DatabaseError |

### `msg_history`

#### Description:

Requests history of messages, up to 50 at a time  
`from` and `to` must be included, and `to` ≥ `from`

#### Request:

``` json
{
    "type": "msg_history",
    "from": 1674007374233575000,
    "to":   1674007406149609000
}

```

#### Response

###### Successful history gathered

``` json
{
    "type": "msg_history",
    "from": 1674007374233575000,
    "to": 1674007406149609000,
    "messages": [
        {
            "MessageID": "63c7534e8cb827613b1e6bef",
            "Content": "Welcome to Fenix!",
            "Timestamp": 1674007374233575000,
            "Author": "63c74c018cb827613b1e6bea"
        },
        {
            "MessageID": "63c753598cb827613b1e6bf0",
            "Content": "Fenix is awesome!",
            "Timestamp": 1674007385580636000,
            "Author": "63c74c018cb827613b1e6bea"
        },
        {
            "MessageID": "63c753678cb827613b1e6bf1",
            "Content": "Making docs isn't fun.",
            "Timestamp": 1674007399333000000,
            "Author": "63c74c018cb827613b1e6bea"
        },
        {
            "MessageID": "63c7536e8cb827613b1e6bf2",
            "Content": "Feeeeeeeenix",
            "Timestamp": 1674007406149609000,
            "Author": "63c74c018cb827613b1e6bea"
        }
    ]
}

```

| **Scenario** | **Response** |
| --- | --- |
| Invalid JSON | JSONDecodeError |
| msg field in msg_send was empty | MessageEmpty |
| Error aggregating messages from database | DatabaseError |
| Invalid from and/or to | Reciprocated request |

## Yodels

### `yodel_create`

#### Description:

Creates a Yodel

#### Request:

``` json
{
    "type": "yodel_create",
    "name": "Fenixland"
}

```

#### Response

###### Successful yodel creation

``` json
{
    "type": "yodel",
    "yodel_id": "63c756d48cb827613b1e6bf3",
    "name": "Fenixland",
}

```

| **Scenario** | **Response** |
| --- | --- |
| Invalid JSON | JSONDecodeError |
| Blank Yodel Name | YodelNameEmpty |
| Error inserting yodel into database | DatabaseError |

### `yodel_get`

#### Description:

Gets a yodel by its ID.

#### Request:

``` json
{
    "type": "yodel_get",
    "yodel_id": "63c756d48cb827613b1e6bf3"
}

```

#### Response

###### Successful

``` json
{
    "type": "yodel",
    "yodel_id": "63c756d48cb827613b1e6bf3",
    "name": "Fenixland",
}

```

| **Scenario** | **Response** |
| --- | --- |
| Invalid JSON | JSONDecodeError |
| Missing ID field | MissingID |
| ID formatted incorrectly | IDFormattingError |
| Yodel specified by ID doesn't exist | YodelDoesntExistError |