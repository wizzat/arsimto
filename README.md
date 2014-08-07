arsimto: A Radically-Simple Inventory Management Tool
=====================================================

Impetus:
========

I looked at Cobbler. I looked at Clusto. I looked at Helix. Every tool seemed overkill for the problem I'm solving.

The problem:
============

1. I have thousands of assets.
2. The assets are grouped in multiple ways.

I want to ask:

1. What are all the assets related to each other?
2. What is some data about the asset?

I want to tie the tool into other tools like Nagios, Ansible, and various dashboards I'm creating (for Grafana, for example). Therefore the tool needs an extremely simple and obvious API that's immediately understood by competent technical professionals.

The solution:
=============

The Unix file system, with its directory structure and symlinking should have the power to express the solution to this problem. A simple wrapper script that adds assets and does "find" commands should suffice to express the solution to the problem.

Assets
======

There is a directory Assets/ that contains one directory per asset. Inside the [asset]/ directory are files named "[variable]:[value]" for example "mac:32:a2:f3:c7:26:6a:20:f9:dd:a5:4e" or "ip:192.168.1.5".

Pools
=====

There is a directory Pools/ that contains one directory per logical grouping. Inside the <pool>/ directory are symlinks to each Assets/[asset]/ that belong to that pool. It is possible for a pool to contain another pool (that is, to be a grouping of pools).

Relationships
=============

A given pool can contain another pool. For example, a data center is a pool. A rack is a pool contained within a data center. Some pools are purely logical, such as "Databases" or "WWW".

Typically we have pools containing assets. For example, the "Databases" pool containing "db01".

Examples / Tutorial
===================

You can get this example output simply by running the tool with no arguments, but it is duplicated here for completeness and for those who want to understand the tool before downloading it.

First, we will create several assets and connect some servers to a switch:

    arsimto add --assets=dc01   --data=capacity:5000
    arsimto add --assets=rack01   --data=U:48
    arsimto add --assets=switch01 --data=ports:48
    arsimto add --assets=server01,server02,server03 --data=ram:16GB,disk:2048GB,nic:10Gb
    arsimto add --assets=server01 --data=ip:192.168.1.101
    arsimto add --assets=server02 --data=ip:192.168.1.102
    arsimto add --assets=server03 --data=ip:192.168.1.103
    arsimto connect --assets=switch01,server01,server02,server03
    
See how things look so far with an "arsimto list" command:

    arsimto ls Pools/switch01/
    server01
    server02
    server03

Try also "arsimto ls Assets" - this will list every asset you have.

Let's connect up those servers to the switch. And while we're at it, add "db" and "www" pools. And why not? Put rack01 into dc01 and switch01 into rack01.
    
    arsimto ln --assets=dc01,rack01
    arsimto ln --assets=rack01,switch01
    arsimto ln --assets=www,server01
    arsimto ln --assets=db,server02,server03

Use "arsimto ls" to see what pools you have.

    arsimto ls Pools/
    db
    dc01
      rack01
        switch01
    www

This shows us the hierarchy of dc01 --> rack01 --> switch01, because when you "connect" a pool to another, they automatically create a hierarchy like this. Note that some objects, like rack01, might be both an asset and a pool. Other things, like "databases" would be only a logical pool.

Let's start in on some reporting. We need a list of all *database* servers with their *IP addresses* and how much *RAM* they have available on them.

    arsimto report  --data=ip,ram  Pools/db
    server02	192.168.1.102	16GB
    server03	192.168.1.103	16GB

Now all the *WWW* servers with their *IP addresses* and *NIC* capacity. Note the order of arguments is unimportant after the initial command-mode argument.
    
    arsimto report Pools/www --data=ip,nic
    server01	192.168.1.101	10Gb

That concludes the examples/tutorial section.

Technical Notes
===============

Pools, although they can be nested, must still be globally unique names! Therefore, if the "Oregon" pool is inside the "AWS" pool, that disallows you from creating an "Oregon" pool inside the "Rackspace" pool. You may choose to put implied hierarchies into your pool names such as "AWS-Oregon" and "Rackspace-Oregon" for example.

You can do a lot of exploration outside the tool. Everything is stored as directories and files in the CWD from where you run the tool. Also note this means two different users will not see the same things if they are in different directories. You can edit the source code of arsimto to make the Pools/ and Assets/ directories be somewhere constant like /opt/arsimto/Pools/ and /opt/arsimto/Assets/, for example.

The Unix filesystem is a tree structure. I am implementing an acyclic graph structure on top of the Unix filesystem by storing Pool names in a directory with symlinks to other Pools and Assets. If you create cycles in the graph (say dc01 :: switch01 :: dc01) then arsimto will be unhappy, and so will you be. If you think you have a case for a cycle in the graph, let's discuss it.

The Unix filesystem has a very well-understood API. Because I'm not storing data in the files, only file names, the API to traverse the data outside of arsimto is "ls", "mv", "find", "rm", and "ln." You will find you don't need "cat" or "grep" commands since the data is stored only in filename metadata.

FAQs
====

 * Q: Backups? A: Try tar -zcvPf backup.tgz /path/to/arsimto/AssetsPoolsDir.
 * Q: Many simultaneous users? A: Try putting AssetsPoolsDir into git. This might also be considered your "backup."
 * Q: Oh noes I deleted half my infrastructure! A: Did you do backups or keep everything in git? Try that.

