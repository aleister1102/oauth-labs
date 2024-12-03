[![License](https://img.shields.io/github/license/cyllective/oauth-labs?_1)](LICENSE)

# OAuth Labs

The main goal of OAuth labs is to learn more about OAuth, its defense
mechanisms and ways to exploit faulty implementations thereof.

The main theme of the labs is to obtain authorization to access the resources
of the `admin` user.

The labs have flags that can be captured once the main vulnerability has been
exploited. Each lab only contains one (known to us) vulnerability which the
user is intended to exploit.

The labs can be solved in either a blackbox or whitebox approach. If the
whitebox approach is chosen, the user should read and analyze the source code
to locate and exploit the vulnerability. Note however that `*.sql` files should
not be read because they contain the flag(s).

For each lab, you should start out by registering an account on the
authorization server first. Once you've set up your account, you can start
digging into the lab.


## Getting started

To get started with the labs, you'll need to have the docker daemon running and
need to tweak your hosts file a little.

**Note:** The lab uses `172.16.16.0/24` as its subnet, ensure this doesn't
collide with your network.

Add the following hosts entries under `/etc/hosts`:

```
172.16.16.1 oauth.labs
172.16.16.1 victim.oauth.labs
172.16.16.1 server-00.oauth.labs server-00
172.16.16.1 client-00.oauth.labs client-00
172.16.16.1 server-01.oauth.labs server-01
172.16.16.1 client-01.oauth.labs client-01
172.16.16.1 server-02.oauth.labs server-02
172.16.16.1 client-02.oauth.labs client-02
172.16.16.1 server-03.oauth.labs server-03
172.16.16.1 client-03.oauth.labs client-03
172.16.16.1 server-04.oauth.labs server-04
172.16.16.1 client-04.oauth.labs client-04
172.16.16.1 server-05.oauth.labs server-05
172.16.16.1 client-05.oauth.labs client-05
```

Once your hosts file is updated, go ahead and invoke the following commands to
clone, build, configure and spawn the labs:

```bash
git clone https://github.com/cyllective/oauth-labs 
cd oauth-labs
make config
make docker
make labs
```

Access the lab index under [https://oauth.labs/](https://oauth.labs/).
Some of the labs require user interaction, you can find a simulator under [https://victim.oauth.labs/](https://victim.oauth.labs/).


## Commands

Below you'll find a list of commonly used commands.

```bash
# Build the docker images
make docker

# Generate configuration files
make config

# Spawn the labs
make labs

# Alternatively, you can spawn individual labs
# make lab00
# make lab01
# make lab02
# make lab03
# make lab04
# make lab05

# Tail docker-compose logs
docker compose -f ./docker-compose.yaml logs -f

# Destroy the labs once you're done
make labsdown
```

## Labs

### Lab 00

This is just a playground, it is used as a base to build new labs by removing
or altering existing code. It shouldn't contain any flags or known
vulnerabilities.

It may be used as a practical way to step through the requests to better
understand the authorization code flow and get a feel for the lab environment.


### Lab 01

+ [server-01.oauth.labs](https://server-01.oauth.labs/)
+ [client-01.oauth.labs](https://client-01.oauth.labs/)

Claims fail; see what happens when a client implementation uses unstable claims
to establish a user identity.


### Lab 02

+ [server-02.oauth.labs](https://server-02.oauth.labs/)
+ [client-02.oauth.labs](https://client-02.oauth.labs/)

Open redirect (No restriction)
See what happens when the authorization server does not validate the
`redirect_uri` at all.

For this lab, [victim.oauth.labs](https://victim.oauth.labs/) can be used to simulate victim interaction.


### Lab 03

+ [server-03.oauth.labs](https://server-03.oauth.labs/)
+ [client-03.oauth.labs](https://client-03.oauth.labs/)

Open redirect (relative path restriction) See what happens when the
authorization server only validates the `redirect_uri` domain.

For this lab, [victim.oauth.labs](https://victim.oauth.labs/) can be used to simulate victim interaction.


### Lab 04

+ [server-04.oauth.labs](https://server-04.oauth.labs/)
+ [client-04.oauth.labs](https://client-04.oauth.labs/)

JWT signature validations are a must, see what happens when they are not verified.


### Lab 05

+ [server-05.oauth.labs](https://server-05.oauth.labs/)
+ [client-05.oauth.labs](https://client-05.oauth.labs/)

JWT signature validations done wrong, see what happens when `jku` claims are
not properly handled.


## Help I'm stuck!

### Callbacks

In case you get stuck on that one callback you just don't receive, ensure
you're using the gateway address `172.16.16.1` instead of localhost. Remember,
the labs are dockerized.

### Use the source, Luke

Reading the source code may help you understand the problem you're having
better, don't shy away from cracking open your code editor and walking through
the code.

### Nope, it's broken!

If all else fails, check back for walkthroughs or reach out. :)


## References

Did you write a blog post, article or refer to oauth-labs in some shape or
form? Add it to our list of [REFERENCES.md](REFERENCES.md) by forking and
opening a Pull request!


## Licensing

[![License](https://img.shields.io/github/license/cyllective/oauth-labs?_1)](LICENSE)

This program is free software: you can redistribute it and/or modify it under the terms of the [MIT license](LICENSE).
