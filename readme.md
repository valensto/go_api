## GOLANG API REST

Hello and welcome to my first little Golang rest api, it's a little api to serve a front end of fish shop to create orders without payment (it's just a preparation orders)

## Motivation

I'm new to Golang and junior to programming (start programming 2 years ago in front end) and wanna to improve my skills. I'm not from backend world but want to go to this direction.
I start Go with framework (echo) and waste lot of time to stay in my confort zone as JS developper and the framework world. I don't say framework isn't good but when you start to learn it's not an option.
I think since I start golang I've better knowledges with programming, reading more book and I really appreciate my programmation change.

**DON'T PRESERVE ME** please say all things you think about my code what I did wrong and what is missing to my api to be ready to go to production !
The project isn't finish, I don't want all answers just a way to make a better app :)

## Code style

I try to make my code without dependence and try to be able to change my dependencies to anothers, I know some things are probably wrong or a least not good but I tried to do my best for this first project. Probably wrote test help me to make it better.
I tried to follow Matt Ryer approch, so please excuse me if I didn't understand think :)

## Tech/framework used

- chi router (I already did a crash project to test routing with only a standard lib)
- mongodb (I'm learning mongodb with this projet)
- jwt-go
- viper
- validator v.9

## Installation

To init dev project

    make run MIGRATE=true DEV=true

To start with migration

    make run MIGRATE=true

To start in dev mode

    make run DEV=true

To start in prod mode

    make run

## Tests

I know I didn't write test and I'm not proud about this I promise I'll write it the next app because testing with postman was sooooo long.

## Credits

[Matt Ryer](https://medium.com/@matryer/how-i-write-go-http-services-after-seven-years-37c208122831)

#### Anything else that seems useful

You can send me your remarks about my code to v.e.brochard@gmail.com or with the golang slack community @valensto

Sorry for my English I'm French, and thank you to read me and thank you to Go community I really appreciate all you do for newbies as me ! Have a nice day :)
