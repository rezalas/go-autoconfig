# Go Autoconfig - A Simple autoconfig service for email servers

Everyone hates email configuration, and server management is no different. Why do server maintainers hate it? Because users need to run clients to connect, and those clients have all created their own standards. Every mail client seems to have a different standard, and if everyone has a standard, nothing is standard.

[![XKCD has a comic for everything](https://imgs.xkcd.com/comics/standards.png)](https://xkcd.com/927/)

Go Autoconfig is a simple autoconfig service designed to integrate well with pretty much any email server, but it is particularly well-suited for Postfix/Dovecot setups. It provides a simple API for clients to query for their email configuration, and it can be easily integrated into existing mail server setups. It is not opinionated, supports multiple domains, and is easy to set up. If you like it and want changes, fork it and submit a pull request. If you want to contribute but don't know how, open an issue and we can discuss it. If you hate it, fork it and do better. I'm not here to judge you.

To my community, haters, and random people viewing this: thank you for your support and contributions. I hope this project can help make email configuration easier for everyone. If you have any questions or suggestions, please feel free to open an issue or submit a pull request.
