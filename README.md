arsimto: A Radically-Simple Inventory Management Tool
=====================================================

Impetus:
========

I looked at Cobbler. I looked at Clusto. I looked at Helix. Every tool seemed
overkill for the problem I'm solving.

The problem:
============

1. I have thousands of assets.
2. The assets are grouped in multiple ways.

I want to ask:

1. What are all the assets related to each other?
2. What is some data about the asset?

I want to tie the tool into other tools like Nagios, Ansible, and various
dashboards I'm creating (for Grafana, for example). Therefore the tool needs an
extremely simple and obvious API that's immediately understood by competent
technical professionals.

The solution:
=============

The Unix file system, with its directory structure and symlinking should have
the power to express the solution to this problem. A simple wrapper script that
adds assets and does "find" commands should suffice to express the solution to
the problem.

Assets
======

There is a directory Assets/ that contains one directory per asset. Inside the
[asset]/ directory are files named "[variable]:[value]" for example
"mac:32:a2:f3:c7:26:6a:20:f9:dd:a5:4e" or "ip:192.168.1.5".

Pools
=====

There is a directory Pools/ that contains one directory per logical grouping.
Inside the <pool>/ directory are symlinks to each Assets/[asset]/ that belong
to that pool. It is possible for a pool to link to another pool (that is, to be
a grouping of pools).

Note that the pools aren't contained within each other. They're merely linked.
The pool namespace is flat. When you do "arsimto ls Pools" the tool attempts to
show you the linking relationships using indentation, which is traditionally
used to show nesting. This can be confusing. If a better UI for indicating
links is discovered, it will be implemented.

Because links are implemented as symlinks within a directory, you can fool
yourself by doing "arsimto ls Pools/GroupingPool/GroupedPool" and it will work.
This helps reinforce the incorrect perception that there is nesting (aka
Parent/Child) but such nesting does not exist. I apologize for this
confusing aspect now.

Relationships
=============

A given pool can link to another pool. For example, a data center is a pool. A
rack is a pool linked from the data center. Some pools are purely logical, such
as "Databases" or "WWW". Others are physical, such as "Rack" or "Switch".

Pools can link to physical assets. For example, the "Databases" pool containing
"db01". Pools can also link to other pools. For example, the "AWS" pool
containing "OR" and "VA" (US-West and US-East names if you prefer the Amazon
naming).

Examples / Tutorial
===================

If you run "arsimto -h" you will get a verbose help output including "Examples"
commands that, when run in order, will create a very simple datacenter. Here we
will do a slightly more-complex real-world setup with more assets. All the
following happens in a Linux BASH shell. Let's begin by creating a number of
assets:

    for i in {01..20} ; do arsimto add --assets=server-$i.or --data=ip:54.0.0.$i ; done
    for i in {01..20} ; do arsimto add --assets=server-$i.sf --data=ip:54.0.1.$i ; done

Now we'll segregate them according to their purpose. We'll have some MySQLs,
some WWWs, and some memcacheds. We'll add machine-specific data and put
them into pools:

    for dc in or sf ; do
	    for i in {01..05} ; do
		    arsimto add --assets=server-$i.$dc --data=disk:8tB,network:20gb,ram:64GB,cpus:24 ;
			arsimto ln --assets=mysql,server-$i.$dc ;
		done ;
	    for i in {06..10} ; do
		    arsimto add --assets=server-$i.$dc --data=disk:1tB,network:10gb,ram:8GB,cpus:8 ;
			arsimto ln --assets=www,server-$i.$dc ;
		done ;
        for i in {11..20} ; do
		    arsimto add --assets=server-$i.$dc --data=disk:8gB,network:40gb,ram:32GB,cpus:4 ;
			arsimto ln --assets=memcached,server-$i.$dc ;
		done ;
	done

Note that if this is real life, you might choose not to do the --data= option
for your add, but rather --collect which will SSH to the server and collect
some data from it. Tying the "add" portion into Facter might be a nice addition
in the future.

Now we have two datacenters, we should reflect that. We'll pretend these are in
Amazon.

    for i in {01..20} ; do arsimto ln --assets=SF,server-$i.sf ; done
    for i in {01..20} ; do arsimto ln --assets=OR,server-$i.or ; done
    arsimto ln --assets=AWS,OR
    arsimto ln --assets=AWS,SF

Now let's see how things look:

    arsimto ls Pools/
    AWS
      OR
      SF
    memcached
    mysql
    www

Remember, the level of indent indicates a relationship, not nesting. OR and SF
are just top-level pools like anything else:

    arsimto ls Pools/OR
    server-01.or
    server-02.or
    server-03.or
    ...[snip]...
    server-20.or

Now let's move some servers around to complicate things, like real life:

    arsimto mv Pools/memcached/server-1[4-6]* Pools/www/
    arsimto mv Pools/www/server-0[6-7]* Pools/mysql/

Now what do things look like?

    arsimto report Pools/www --data=ip,disk,ram,network | fgrep .or
    server-08.or	54.0.0.08	1tB	8GB	10gb
    server-09.or	54.0.0.09	1tB	8GB	10gb
    server-10.or	54.0.0.10	1tB	8GB	10gb
    server-14.or	54.0.0.14	8gB	32GB	40gb
    server-15.or	54.0.0.15	8gB	32GB	40gb
    server-16.or	54.0.0.16	8gB	32GB	40gb

This is where an inventory management system starts to help matters. We've
added chaos into our system, and this helps us keep track of the chaos. Our WWW
pool is heterogenous, so we might weight the servers differently, or start
doubling up processes.

Technical Notes
===============

Pools, although they can be nested, must still be globally unique names!
Therefore, if the "Oregon" pool is inside the "AWS" pool, that disallows you
from creating an "Oregon" pool inside the "Rackspace" pool. You may choose to
put implied hierarchies into your pool names such as "AWS-Oregon" and
"Rackspace-Oregon" for example.

You can do a lot of exploration outside the tool. Everything is stored as
directories and files in the CWD from where you run the tool. It's suggested to
keep these files in git or subversion so that users don't clobber each others'
changes, and you don't have to worry about permissions of shared directories.

The Unix filesystem is a tree structure. I am implementing an acyclic graph
structure on top of the Unix filesystem by storing Pool names in a directory
with symlinks to other Pools and Assets. If you create cycles in the graph (say
dc01 :: switch01 :: dc01) then arsimto will be unhappy, and so will you
be. If you think you have a case for a cycle in the graph, let's discuss it.

The Unix filesystem has a very well-understood API. Because I'm not storing
data in the files, only file names, the API to traverse the data outside of
arsimto is "ls", "mv", "find", "rm", and "ln." You will likely be doing "sort"
and "grep" commands on the output reports occasionally.

FAQs
====

 * Q: Backups? A: Try tar -zcvPf backup.tgz /path/to/arsimto/AssetsPoolsDir.
 * Q: Many simultaneous users? A: Try putting AssetsPoolsDir into git. This might also be considered your "backup."
 * Q: Oh noes I deleted half my infrastructure! A: Did you do backups or keep everything in git? Try that.

