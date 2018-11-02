Login Openshift.io users
========================

An utility to login Openshift.io users and get auth and refresh tokens.

Prerequisities
--------------

Chrome or [Chromium browser](https://www.chromium.org/Home) with headless feature and [Chromedriver](https://sites.google.com/a/chromium.org/chromedriver/) needs to be installed where it is run (for Fedora/RHEL/CentOS):

```shell
sudo yum install chromium chromium-headless chromedriver
```

Usage
-----

To run, provide a line separated list of users ("user=password") in the property file defined by the `USERS_PROPERTIES_FILE` environment variable and execute:

```shell
go run loginusers_oauth2.go
```

Configuration via environment variables:

* `AUTH_SERVER_ADDRESS` = server of Auth Services (default `https://auth.openshift.io`).
* `AUTH_CLIENT_ID` = client id (default `740650a2-9c44-4db5-b067-a3d1b2cd2d01`).
* `USERS_PROPERTIES_FILE` = a file containing a line separated list of users in a form of `user=password` (default `users.properties`).
* `USER_TOKENS_FILE` = an output file where the generated auth and refresh tokens were written after succesfull login of each user (default `user.tokens`).
* `USER_TOKENS_INCLUDE_USERNAME` = "`true` if username is to be included in the output (default `talse`).
* `MAX_USERS` = A maximal number of users taken from the `USERS_PROPERTIES_FILE` (default `-1` means unlimited).

Example:

```shell
AUTH_SERVER_ADDRESS=https://auth.prod-preview.openshift.io -Duser.tokens.file=osioperftest.tokens go run loginusers_oauth2.go
```
