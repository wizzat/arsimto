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
adds assets and does various commands should suffice to express the solution to
the problem.

So far, the list of Unix commands I've needed to achieve the goal:

 * find
 * ls
 * mv
 * ln
 * rm

Assets
======

There is a directory Assets/ that contains one directory per asset. Inside are
files named "[variable]:[value]" for example:

 * mac:32:a2:f3:c7:26:6a:20:f9:dd:a5:4e
 * ip:192.168.1.5
 * ram:32GB
 * cpus:8

As you can see, values can have colons. The variable name cannot. Everything up
to the first colon is the variable name (which you can report on). Try not to
be creative with these. Don't use shell metacharacters, like any sort of
punctuation.

Pools
=====

There is a directory Pools/ that contains one directory per logical grouping.
Inside are symlinks to each object that belongs to that Pool. Typically that
object is an Asset. However, you can also aggregate several Pools underneath
another like this:

    arsimto ln mainapp memcached mysql www
    arsimto ls Pools/
    mainapp --> memcached,mysql,www
    memcached
    mysql
    www

The Pool namespace is flat. The Pools aren't contained within each other.
They're merely aggregated by the linking Pool. In precisely the same way an
Asset can appear in two Pools (Production and Databases and Oregon, for
example), a Pool can appear in two other Pools.

The purpose of Pools is to avoid having to type all the names of every Asset or
Pool that are logically grouped every time. 

Examples / Tutorial
===================

If you run "arsimto -h" you will get a verbose help output including "Examples"
commands that, when run in order, will create a very simple datacenter. Here we
will do a slightly more-complex real-world setup with more assets. All the
following happens in a Linux BASH shell. Let's begin by creating a number of
assets:

    for i in {01..20} ; do arsimto add server-$i.or --data=ip:10.2.0.$i ; done
    for i in {01..20} ; do arsimto add server-$i.sf --data=ip:10.2.1.$i ; done

Now we'll segregate them according to their purpose. We'll have some MySQLs,
some WWWs, and some memcacheds. We'll add machine-specific data and put
them into pools:

    for dc in or sf ; do
	    for i in {01..05} ; do
		    arsimto add server-$i.$dc --data=disk:8tB,network:20gb,ram:64GB,cpus:24 ;
			arsimto ln mysql server-$i.$dc ;
		done ;
	    for i in {06..10} ; do
		    arsimto add server-$i.$dc --data=disk:1tB,network:10gb,ram:8GB,cpus:8 ;
			arsimto ln www server-$i.$dc ;
		done ;
        for i in {11..20} ; do
		    arsimto add server-$i.$dc --data=disk:8gB,network:40gb,ram:32GB,cpus:4 ;
			arsimto ln memcached server-$i.$dc ;
		done ;
	done

Note that if this is real life, you might choose not to do the --data= option
for your add, but rather --collect which will SSH to the server and collect
some data from it. Tying the "add" portion into Facter might be a nice addition
in the future.

Now we have two datacenters, we should reflect that. We'll pretend these are in
Amazon.

    for i in {01..20} ; do arsimto ln SF server-$i.sf ; done
    for i in {01..20} ; do arsimto ln OR server-$i.or ; done
    arsimto ln AWS OR
    arsimto ln AWS SF

Now let's see how things look:

    arsimto ls Pools/
    AWS --> OR,SF
    OR
    SF
    memcached
    mysql
    www

Now let's move some servers around to complicate things, like real life:

    arsimto mv Pools/memcached/server-1[4-6]* Pools/www/
    arsimto mv Pools/www/server-0[6-7]* Pools/mysql/

Now what do things look like?

    arsimto report Pools/www --data=ip,disk,ram,network | fgrep .or
    server-08.or	10.2.0.08	1tB	8GB	10gb
    server-09.or	10.2.0.09	1tB	8GB	10gb
    server-10.or	10.2.0.10	1tB	8GB	10gb
    server-14.or	10.2.0.14	8gB	32GB	40gb
    server-15.or	10.2.0.15	8gB	32GB	40gb
    server-16.or	10.2.0.16	8gB	32GB	40gb

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

Q: Backups?
A: Try tar -zcvPf backup.tgz /path/to/arsimto/AssetsPoolsDir.

Q: Many simultaneous users?
A: Try putting AssetsPoolsDir into git. This might also be considered your "backup."

Q: Oh noes I deleted half my infrastructure!
A: Did you do backups or keep everything in git? Try that.

Q: Pool nesting hierarchies?
A: No. Pool namespace is flat. See this example:

    arsimto ls Pools/
	AWS --> OR,SF
	OR
	Rackspace --> OR
	SF
	memcached
	mysql
	www

AWS points to the OR (and SF) pool, but so does Rackspace. This doesn't mean
there are distinct AWS/OR and Rackspace/OR pools. It means they both point to
the same thing. This is almost certainly an error.

Because links are implemented as symlinks within a directory, you can fool
yourself by doing "arsimto ls Pools/GroupingPool/GroupedPool" and it will work.
This helps reinforce the incorrect perception that there is nesting (aka
Parent/Child) but such nesting does not exist. I apologize for this
confusing aspect now.

