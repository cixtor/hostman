### Hostman (Hosts Manager)

Hosts manager with focus on simplicity and support for adword blocker lists. You can add, remove, disable, enable entries in the hosts file; works better with _dnsmasq_ for advanced control of the DNS A record connections and implicit configuration for wildcard domains.

The hosts file is a computer file used by an operating system to map hostnames to IP addresses. The hosts file is a plain text file, and is conventionally named hosts. The DNS implementation automated the publication process and provided instantaneous and dynamic hostname resolution in the rapidly growing network. In modern operating systems, the hosts file remains an alternative name resolution mechanism, configurable often as part of facilities such as the _Name Service Switch_ as either the primary method or as a fallback method.

In some operating systems, the contents of the hosts file is used preferentially to other name resolution methods, such as the DNS, but many systems implement name service switches to provide customization. Unlike remote DNS resolvers, the hosts file is under the direct control of the local computer's administrator.

More info at [WikiPedia Hosts File](https://en.wikipedia.org/wiki/Hosts_(file))

### Usage

When the program is executed it reads the entire hosts file and parses its entries to separate the IP address, the domain, and the possible aliases; additionally it detects if the line is commented and assigns a special flag to say that the entry is currently disabled. You can export the entire file in JSON format using the _export_ command or export only the entries that match certain text using the _search_ command like this:

```
$ hostman -search example
$ hostman -search example -export
```

The program uses the `/etc/hosts` as the default file path to operate all the actions available, if you have multiple configuration files or want to test the commands with a dummy file you can use the _config_ command to force the program to read a different host file like this:

```
$ hostman -config /tmp/hosts -export
$ hostman -config /tmp/hosts -search example
$ hostman -config /tmp/hosts -search example -export
```

Adding new entries to the hosts file requires a special format, a new entry needs to have the IP address followed by the `@` symbol and the domain that is going to be associated to it. Note that the address accepts a wide range of characters to support IPv6 and it does not checks the validity of the host in any point, this is intentional as the connection to certain IP blocks will respond with server failures as they are not properly routed by the DNS server; you must check the integrity of the IP and TLD by yourself. Additionally you can append domain aliases separating them with commas like this:

```
$ hostman -add ff02::5@example.com
$ hostman -add 127.0.0.1@example.com
$ hostman -add 127.0.0.1@example.com,example.org
$ hostman -add 127.0.0.1@example.com,example.org,example.net
```

The removal of host entries requires the execution of the search command, this is to avoid unexpected modifications as the user can inspect the entries that will be affected before the program overrides the file. The program will remove all the entries found with a simple string matching function, if you want to be more specific with the results add more characters like this:

```
$ hostman -search example -remove
$ hostman -search example.com -remove
$ hostman -search foobar.example.com -remove
```

Note that the program does not creates backups before the removal and/or addition of one or more entries, you have to be careful with the data that is added to the hosts file as it may break your Internet connection depending on the configuration of your personal computer or server. If you want to delete a host entry temporarily it is suggested to use the _disable_ command instead which will comment the entries that match the query from the search results, and the _enable_ command to uncomment the same entries like this:

```
$ hostman -search 127.0.0.1 -enable
$ hostman -search 127.0.0.1 -disable
$ hostman -search example.com -disable
```

### Permission Denied

Some of the operations executed by the program require write permissions on the hosts file depending on its location, you may need to suffix some of the commands with _sudo_ or add your current user account to the admin group. Alternatively you can add an alias that executes sudo automatically for every command like this:

```
alias hostman='sudo env "PATH=$PATH" hostman $@'
```

### Use Cases

- Basic management of hosts settings
- Custom TLD domains for local development
- Private connection to non-public jailed IP addresses
- Automate parental control for school and/or study time
- Automated blocking of obstructive websites during office hours

### License

```
The MIT License (MIT)

Copyright (c) 2015 CIXTOR

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
