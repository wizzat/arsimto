arsimto: A Radically-Simple Inventory Management Tool
=====================================================

Impetus:
========

I looked at Cobbler. I looked at Clusto. Every tool seemed overkill for the problem I'm solving.

The problem:
============

1. I have thousands of assets.
2. The assets are related in multiple ways.
3. What are all the assets related to each other?
4. What are some of the attributes around the asset?

The solution:
=============

The Unix file system, with its directory structure and symlinking should have the power to express the solution to this problem. A simple wrapper script that adds assets and does "find" commands should suffice to express the solution to the problem.

Assets
======

There is a directory Assets/ that contains one directory per asset. Inside the [asset]/ directory are files named "[variable]:[value]" for example "mac:32:a2:f3:c7:26:6a:20:f9:dd:a5:4e" or "ip:192.168.1.5".

Pools
=====

There is a directory Pools/ that contains one directory per logical grouping. Inside the <pool>/ directory are symlinks to each Assets/[asset]/ that belong to that pool.

Relationships
=============

A given pool should be able to contain another pool. For example, a data center is a pool. A rack is a pool contained within a data center. A switch is a pool contained within a rack. A pool can also be an asset, in the case of a switch, for example. Or, if you're the data center owner, a rack, an aisle, a cage, etc.

Examples
========

You can get this example output simply by running the tool with no arguments, but it is duplicated here for completeness and for those who want to understand the tool before downloading it.

    arsimto add --assets=switch01 --attrs=ports:48
    arsimto add --assets=server01,server02,server03 --attrs=ram:16GB,disk:2048GB,nic:10Gb
    arsimto add --assets=server01 --attrs=ip:192.168.1.101
    arsimto add --assets=server02 --attrs=ip:192.168.1.102
    arsimto add --assets=server03 --attrs=ip:192.168.1.103
    
Now you've created a basic switch01 with 3 servers attached. You can see how things look so far with some "arsimto list" commands:

    arsimto list Pools/switch01/
    server01
    server02
    server03

Try also "arsimto list Assets" - this will list every asset you have.

Let's connect up those servers to the switch. And while we're at it, add "db" and "www" pools.
    
    arsimto connect --assets=switch01,server01,server02,server03
    arsimto connect --assets=www,server01
    arsimto connect --assets=db,server02,server03

Use "arsimto list" to see what pools you have.

    arsimto list Pools/
    db
    switch01
    www

Let's start in on some reporting. We need a list of all database servers with their IP addresses and how much RAM they have available on them.

    arsimto report Pools/db --attrs=ip,ram
    server02	192.168.1.102	16GB
    server03	192.168.1.103	16GB

Now we want to know about all the WWW servers with their IP addresses and NIC capacity.
    
    arsimto report Pools/www --attrs=ip,nic
    server01	192.168.1.101	10Gb

That concludes the examples/tutorial section.
