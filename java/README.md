Login Openshift.io users
========================

An utility to login Openshift.io users and get auth and refresh tokens.

Prerequisities
--------------

Chrome or [Chromium browser](https://www.chromium.org/Home) with headless feature and [Chromedriver](https://sites.google.com/a/chromium.org/chromedriver/) needs to be installed where it is run (for Fedora/RHEL/CentOS):
```
$ sudo yum install chromium chromium-headless chromedriver
```

Usage
-----

To run, provide a line separated list of users ("user=password") in the property file found at `src/main/resources/users.properties` and execute:
```
$ mvn clean compile exec:java (-Dauth.server.address=...) (-Duser.tokens.file=...)
```

where:
 * `auth.server.address` = server of Auth Services (e.g. "`https://auth.prod-preview.openshift.io`")
 * `user.tokens.file` = an output file where the generated auth and refresh tokens were written after succesfull login of each user

Example:
```
$ mvn clean compile exec:java -Dauth.server.address=https://auth.prod-preview.openshift.io -Duser.tokens.file=osioperftest.tokens
```
