---
title: FastForward Server API Reference

language_tabs: # must be one of https://git.io/vQNgJ
  - Examples

toc_footers:
  - <a href='https://github.com/slatedocs/slate'>Documentation Powered by Slate</a>

includes:
  - errors

search: true

---
# Introduction

Version 1.0.0

Welcome to the FastForward Server API docs! It is recommended to use [Hoppscotch]https://hoppscotch.io/) or [Insomnia](https://insomnia.rest/download) to get familiar with the api before developing.

# Crowd Query

> Query for "https://shortlinksite.example/9njv3"

```
domain=shortlinksite.example&path=9njv3
```
> will return text/plain response

```
asitenotrelatedtopiracy.example
```
For Crowd Query send a **`application/x-www-form-urlencoded` POST** request with the domain and path to **`crowd.fastforward.team/crowd/query_v1`**

<aside class="notice">
Note how the protocol(https://) and the first "/" are not included in the form
</aside>

Response status codes

Staus Code | Response | Description
---------- | -------- | -----------
200        | text/plain | The link was found in the db and returned
204        | empty    | Not found in db

There is a 10% chance that the server will return 201 even if the link was found, forcing the user to go through the site and verify the target.

# Crowd Contribute

> Contribute destination "https://asitenotrelatedtopiracy.example" for "https://shortlinksite.example/9njv3"

```
POST crowd.fastforward.team/crowd/query_v1

domain=shortlinksite.example&path=9njv3&target=asitenotrelatedtopiracy.example
```
> Will return:

```
201 Created
```

For Crowd Contribute send a **`application/x-www-form-urlencoded` POST** request with the domain, path and target to **`crowd.fastforward.team/crowd/contribute_v1`**


### Verifying targets
If the domain and path sent in a crowd contribute request already exist in the db and:
- the target matches the one in the db, nothing happens and status 201 is returned
- the target does not match the one in the db, the entry gets `reported` and status 201 is returned

# Admin endpoint

## Sign up
To create an admin account access the server database and add a row in the `admin_creds` table with the username and [password(hashed with bcrypt cost 10)](https://gchq.github.io/CyberChef/#recipe=Bcrypt(10))

```SQL
INSERT INTO admin_creds (username, password) VALUES ('thebestadmin', '$2a$10$Vae9jdPSR.teZHW5qh0HCebhzVUhH7vWdfSAQd8kdHO/iGPzY8WJS')
```

## Change password

> Request:

```json
POST crowd.fastforward.team/admin/api/changepassword

{
	"username": "thebestadmin",
	"oldpassword": "agoodpassword",
	"newpassword": "areallygoodpassword"
}
```
> will return

```
200 OK
```

To change your password send a **`JSON` POST** request with the username, oldpassword and newpassword keys to 
**`crowd.fastforward.team/admin/api/changepassword`**

Response status codes

Staus Code | Response | Description
---------- | -------- | -----------
200        | empty    | Password was successfully updated
401        | empty    | username not found or old password did not match

<aside class="notice">Changing a user's password will resest their refresh token</aside>

## Getting refresh tokens

> Getting a refresh token for thebestadmin

```json
POST crowd.fastforward.team/admin/api/newreftoken`
{
	"username": "thebestadmin",
	"password": "areallygoodpassword"
}
```
> Response

```json
{
  "reftoken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0aGViZXN0YWRtaW4iLCJleHAiOjEwMDMwOTA0OTksImp0aSI6ImQzZjY4ODZkLTcwNmUtNGFhMC05MTZiLWM5ODVjODZhNWFkOSIsInN1YiI6InJlZl90b2tlbiJ9.vSpqFOeHTi7F4rIC9l-D-vxicZOetHvUqa1xJtQMn2OGhz_HlLuIuhsg4rYspAVrhYU-xTlcequW35VjdzR0gKO8OIaGA3CVpCg6PlYqRMXOakbVIdES35xMIGHgHp-_XVGXyZ34htkI5yAp6MI3p1E6oCD-wRdvxq9eRT0u9PjvF1CBw9YsWkUlR_5VS-AyeF8asvEqzwiq9ZRljYKbyEzbqJn-vcb1S7SU4PqSpKgxL9DriL2oC0QqO3N56Lx8gLXszbiSngveBlREzM5XczF8Ii6ap8JjfRwLjXVGmap2fgYzuJEoExIV7G7dIJkU83j3XVt7DEzus5eA6mIMEA"
}

```

To reduce the amount of database lookups the server uses refresh and access tokens. An access token is valid for only 15 minutes and is verified quickly, without a db lookup. However, to generate an access token a refresh token needs to be verified which does need a db lookup. This way only one lookup every 15 minutes is needed for each user.

<aside class="warning">Only one refresh token per user can be generated, if a new one is generated the old one becomes invalid</aside>

A refresh token is valid for 90 days, just ask the user to login again when it expires.

To change your password send a **`JSON` POST** request with the username and password keys to 
**`crowd.fastforward.team/admin/api/newreftoken`**
The server will send a JSON response with a reftoken key

Staus Code | Response         | Description
---------- | ---------------- | -----------
200        | application/json | success
401        | empty            | username not found or password did not match

## Getting access tokens

> Generating an access token

```json
POST crowd.fastforward.team/admin/api/newacctoken

{
  "reftoken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0aGViZXN0YWRtaW4iLCJleHAiOjEwMDMwOTA0OTksImp0aSI6ImQzZjY4ODZkLTcwNmUtNGFhMC05MTZiLWM5ODVjODZhNWFkOSIsInN1YiI6InJlZl90b2tlbiJ9.vSpqFOeHTi7F4rIC9l-D-vxicZOetHvUqa1xJtQMn2OGhz_HlLuIuhsg4rYspAVrhYU-xTlcequW35VjdzR0gKO8OIaGA3CVpCg6PlYqRMXOakbVIdES35xMIGHgHp-_XVGXyZ34htkI5yAp6MI3p1E6oCD-wRdvxq9eRT0u9PjvF1CBw9YsWkUlR_5VS-AyeF8asvEqzwiq9ZRljYKbyEzbqJn-vcb1S7SU4PqSpKgxL9DriL2oC0QqO3N56Lx8gLXszbiSngveBlREzM5XczF8Ii6ap8JjfRwLjXVGmap2fgYzuJEoExIV7G7dIJkU83j3XVt7DEzus5eA6mIMEA"
}
```

> Will return:

```json
{
  "acctoken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0aGViZXN0YWRtaW4iLCJleHAiOjEwMDc5MDg0MTAsInN1YiI6ImFjY190b2tlbiJ9.X5R3VjrxEAxCNp_wLnQwRrFdq6UszPV-B-IDlVrYOzreSgpi65Hmyzh-J7kBx2RRz0NL2WZFiOv8vkQvNAjZQlstogIhoNvFbK8a8FshkimFf-sUYybh7hTzl3JvCAqwoIQ0Bom_TVHUlw988UCEZJxZZVLXcUWM_L9507g12kH9HGAfcsRvjGKpiMhUfystWoxFhLLvutkZBvWoaQ9NxNr0I8_AS5pyBPAi5h6H0RZMhn3zcULEJkOJ1suwjBnnf8MReOGHGmqKLlYadwxjy6iei98fL5l1n2kQkpQreBVyofRrPsWgYwnCRcmiynqLWJ4FzDKb2ksCMQ_eiQCexw"
}
```

<aside class="notice">Access tokens are valid for only 15 minute remember to generate and use new ones often</aside>

To change your password send a **`JSON` POST** request with a reftoken key to 
**`crowd.fastforward.team/admin/api/newacctoken`**
The server will send a JSON response with an acctoken key

Staus Code | Response         | Description
---------- | ---------------- | -----------
200        | application/json | success
401        | empty            | invalid reftoken

## Getting reported links

> Request

```
> POST /admin/api/getreported

> Content-Type: application/x-www-form-urlencoded

> Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0aGViZXN0YWRtaW4iLCJleHAiOjE2Mzc5MDk2NjUsInN1YiI6ImFjY190b2tlbiJ9.m2PZ8Lv6KckQEC6lw3g655BgY0z1XcaeqxxGnfuP0YE-h7aU6YUbFyOAjPZedxLLhNzAcDZp806lFBAAiDGiEz79Jz_YYrISDSlb7YiwHSbarSvMZLnBOMkxbcsVAoA45Pqn0u_uHRhyeTr8d7r5dlMASZv8AUnEa_UkjfArOwRrv_hb0boC_AlLP5mL3NyGfHbVAPza7Jc9jyxCf46uvoSsR5X6zi17ROOFYytkfzUeyk6HiJ_nQ6wCfWqbDkh1cdTqsKPcCWB141UtmoNPAfqxhT-zK-MLkGGKFLQJIP0_xMYTVwB-Mya3tIFmm-CQsyPFGsy6H5LGRAsXOcRYew

| page=1

```

> Response

```json
[
  {
    "id": 3,
    "domian": "shortlinksite.example",
    "path": "9njv3",
    "destination": "fishylink.example",
    "times_reported": 1,
    "hashed_IP": "64ced7e32efb08f35afccadca888750d23263a5850e62cc8846052bc93ceef7c",
    "votedfordeletion": false,
    "voted_by": ""
  },
  {
    "id": 5,
    "domian": "shortlinksite.example",
    "path": "9mj58",
    "destination": "suslink.example",
    "times_reported": 8,
    "hashed_IP": "61be55a8e2f6b4e172338bddf184d6dbee29c98853e0a0485ecee7f27b9af0b4",
    "votedfordeletion": false,
    "voted_by": ""
  },
  {
    "id": 6,
    "domian": "shortlinksite.example",
    "path": "a8dyu123",
    "destination": "imposterlink.example",
    "times_reported": 1,
    "hashed_IP": "f9cf2e767495e184a9d07523a3cd9f18faf1b4975ed632f9072009a03e1774e8",
    "votedfordeletion": true,
    "voted_by": "thebetteradmin"
  }
]
```

To get reported links send a **`form-urlencoded` POST** request with a page parameter and access token as bearer token in the Authorization header to 
**`crowd.fastforward.team/admin/api/getreported`**

This will return up to 20 entries at once. Use the page argument to get another set of 20 entries.

Staus Code | Response         | Description
---------- | ---------------- | -----------
200        | application/json | success
401        | empty            | bearer token in auth header is invalid

## Vote for deletion

> Request

```json
{
	"domain": "shortlinksite.example",
	"path": "a8dyu123"
}
```
> Response

```go
empty

```
 
Admins can vote to delete an entry. It takes 2 votes to delete an entry. *The server does not prevent an admin from voting twice.* For now, only the the first voter is stored in the db (Both voters do show up in logs, though).

To vote send a **`JSON` POST** request with keys for the domain and path with the access token as bearer token in the Authorization header to
**`crowd.fastforward.team/api/votedelete`**

Staus Code | Response         | Description
---------- | ---------------- | -----------
200        | empty            | Successfully voted (The entry now has a total of `1` vote)
202        | empty            | Successfully deleted, the entry already had one vote
422        | empty            | domain and path not found
