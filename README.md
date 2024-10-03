# coda-hiring-exercise
Hiring exercise coda shop/payments

HTTP Round Robin API

Goal: Write a Round Robin API which receives HTTP POSTS and routes them to one of a list of
Application APIs

CREATE A SIMPLE API
Create a simple APPLICATION API with one endpoint which accepts an HTTP POST with any
JSON payload. The API will respond with a successful response containing an exact copy of
the JSON request it received. The request and response structure can be any valid JSON, but
below is a good example to follow.
For example if you post:

{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}

Your API should respond with an HTTP 200 success code with a response payload of:

{"game":"Mobile Legends", "gamerID":"GYUTDTE", "points":20}

You should be able to run multiple instances of this API (for example on different ports) - for
your demo you should have at least 3 instances.

CREATE A ROUTING API
Create an ROUND ROBIN API which will receive HTTP POSTS and send them to an instance of
your application API. The round robin API will receive the response from the application API
and send it back to the client.
You should be able to configure the round robin API with a list of application API instances e.g.
if you run 3 instances of the application API then the round robin API will need the addresses of
these instances.


When the round robin API receives a request it should choose which application API instance to
send the request to on a ‘round robin’ basis. Therefore if you have 3 instances of the
application API then the first request goes to instance 1, the second to instance 2, the third to
instance 3 etc
Please code the logic for round robin yourself and don't rely on an external framework or library
to provide this functionality for you. It's ok to use frameworks to create the HTTP API and
service basics, but for round robin logic we would like you to code this yourself.
These are the basic requirements of the code, but here are some things for you to consider:
● How would my round robin API handle it if one of the application APIs goes down?
● How would my round robin API handle it if one of the application APIs starts to go
slowly?
● How would I test this application?
You may write the code in any language and then share the code with us before the demo
(git/bitbucket/zip file are all fine - we just want a chance to review your code before the
interview). During the demo you will share your screen and review the code live in your IDE
with our interviewers.
We want to be respectful of your time so please time-box how much effort to put into this and
most of all have fun writing the code!
