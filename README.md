arsimto: A Radically-Simple Inventory Management Tool
=====================================================

Impetus
=======

I looked at OpenDCIM, Cobbler, Collins, Clusto, Helix. Every tool seemed
difficult or overkill for the problem I'm solving.

The Problem
===========

1. I have thousands of assets.
2. The assets are grouped in multiple ways.

I want to ask:

1. What are all the assets related to each other?
2. What is some data about the asset?

A concrete example, since the above might be difficult to understand: I want
a list of all Production Database servers in Oregon along with their IP
addresses and total RAM configured on the system.

This would help me decide what further scripts should be run on that
intersection to tie the output of the tool into other tools like Nagios,
Ansible, and various dashboards I'm creating (for Grafana, for example).
Therefore the tool needs an extremely simple and obvious API that's immediately
understood by competent technical professionals.

The Solution
============

The Unix file system combined with a simple wrapper script should have the
power to express the solution to this problem.

The most damning limitation of this solution is the inability to use the
special character `/` in any data value. The ipv6 address and ipv4 netmask
are two data values that must be re-encoded. I use the other special
character `\` to encode it, which means you cannot use `\` in your data
values.

There is also a helper script, arsimtoREST which will listen on a port and
transfer requests to arsimto. So for example, you might do this:

    [inventoryhost]$ ./arsimtoREST -p=12345 >logfile 2>&1 &
    [anotherhost]$ curl 'http://my.inventoryhost.localdomain:12345/arsimto/ls/-l/-j/-i/MySQLs/SF/-d=name' 2>/dev/null

Which will return this (all MySQL servers in SF):

    {
        "arsimtoAssets":[
          { "name":"srvr1a.sf" }
        , { "name":"srvr1b.sf" }
        , { "name":"srvr1c.sf" }
        , { "name":"srvr2a.sf" }
        , { "name":"srvr2b.sf" }
        , { "name":"srvr2c.sf" }
        , { "name":"srvr3a.sf" }
        , { "name":"srvr3b.sf" }
        , { "name":"srvr3c.sf" }
        ]
    }

See the section below on arsimtoREST for more help on setting it up.

Assets
======

There is a directory Assets/ that contains one directory per asset. Inside are
files named `[variable]:[value]` for example:

 * cpus:8
 * dnsname:hodor.foo.tld
 * eth0-ip:192.168.1.5
 * eth1-ip:10.0.0.1
 * eth0-mac:32:a2:f3:c7:20:f9
 * eth1-mac:32:a2:f3:c7:2f:5f
 * intip:10.0.0.1
 * ip:192.168.1.5
 * memkB:16777216
 * sda1-diskMB:12582912
 * sdb1-diskMB:1098907648

As you can see, values can have colons. The variable name cannot. Everything up
to the first colon is the variable name (which you can report on). Try not to
be creative with variable names. Don't use shell metacharacters.

Pools
=====

There is a directory Pools/ that contains one directory per logical grouping.
Inside are directories corresponding to each object that belongs to that Pool.
Typically that object is an Asset. However, you can also aggregate several
Pools underneath another like this:

    arsimto ln mainapp memcached mysql www
    arsimto ls -l Pools/
    (mainapp) --> (memcached) (mysql) (www)
    (memcached) --> +++
    (mysql) --> +++
    (www) --> +++

The Pool namespace is flat. The Pools aren't contained within each other.
They're merely aggregated by the linking Pool. In precisely the same way an
Asset can appear in two Pools (Production and Databases and Oregon, for
example), a Pool can appear in two other Pools.

The purpose of Pools is to avoid having to type all the names of every Asset or
Pool that are logically grouped every time.

The arsimtoREST Helper
======================

 1. Put arsimto into the `$PATH` of the user executing arsimtoREST.
 2. Put `$HOME/.arsimtorc` for the user executing arsimtoREST.
 3. Add `--top=/path/to/arsimto/inventory/` into the .arsimtorc.
 4. Execute `./arsimtoREST --port=54321 >logfile 2>&1 &`.
 5. Submit an issue if this doesn't work.

Examples
========

Generate DNS records for all the Staging Couchbase servers:

    $ arsimto ls -i Pools/Couchbases/ Pools/Staging/ -d='name,"IN","CNAME",ip'
    couch-1.bit	IN	CNAME	192.168.0.5
    couch-2.bit	IN	CNAME	192.168.0.1
    couch-3.bit	IN	CNAME	192.168.0.3

Generate Ansible hosts file for all Production MySQL servers:

    $ echo "[RunOnTheseHosts]" ; \
		arsimto ls -p -i Pools/Production/ Pools/MySQLs/ -d=ip,memkB,dnsname \
		| awk '{print $1"   mem="$2"   dnsname="$3}'
	[RunOnTheseHosts]
	192.168.1.212     mem=70196916   dnsname=mysql1a.dc.tld
	192.168.96.109    mem=70196916   dnsname=mysql1b.dc.tld
	192.168.96.125    mem=70196916   dnsname=mysql1c.dc.tld
	192.168.254.152   mem=70196916   dnsname=mysql2a.dc.tld
	192.168.121.239   mem=70196916   dnsname=mysql2b.dc.tld
	192.168.18.82     mem=70196916   dnsname=mysql2c.dc.tld
	192.168.254.206   mem=70197168   dnsname=mysql3a.dc.tld
	192.168.254.149   mem=70197168   dnsname=mysql3b.dc.tld
	192.168.254.139   mem=70197168   dnsname=mysql3c.dc.tld

Re-collect all data for a given set of hosts (maybe you replaced some hardware).
Remove the final `| sh` portion if you want to see what it would do without actually
doing it:

	arsimto ls -i Pools/Production/ Pools/Memcached/ -d=name,ip \
		| awk '{print "arsimto add "$1" --collect="$2}' \
		| sh

Change the hostname of all hosts in a Pool to match the inventory:

	arsimto ls Pools/Memcached/ -d=ip,dnsname \
	   | awk '{print "echo "$1" ; ssh "$1" \"sudo hostname "$2"\""}' \
	   | sh

Tutorial
========

If you run `arsimto -h` you will get a verbose help output including "Examples"
commands that, when run in order, will create a very simple datacenter. Here we
will do a slightly more-complex real-world setup with more assets. All the
following happens in a Linux BASH (things might be a bit odd on OSX) shell.
Let's begin by creating a number of assets:

    for i in {101..120} ; do arsimto add server-$i.or --data=ip:10.2.0.$i ; done
    for i in {101..120} ; do arsimto add server-$i.sf --data=ip:10.2.1.$i ; done

Now we'll segregate them according to their purpose. We'll have some MySQLs,
some WWWs, and some memcacheds. We'll add machine-specific data and put
them into pools:

    for dc in or sf ; do
        for i in {101..105} ; do
		    arsimto add server-$i.$dc --data=disk:8tB,network:20gb,ram:64GB,cpus:24 ;
			arsimto ln mysql server-$i.$dc ;
		done ;
	    for i in {106..110} ; do
		    arsimto add server-$i.$dc --data=disk:1tB,network:10gb,ram:8GB,cpus:8 ;
			arsimto ln www server-$i.$dc ;
		done ;
        for i in {111..120} ; do
		    arsimto add server-$i.$dc --data=disk:8gB,network:40gb,ram:32GB,cpus:4 ;
			arsimto ln memcached server-$i.$dc ;
		done ;
	done

Note that if this is real life, you might choose not to do the `--data=` option
for your add, but rather `--collect` which will SSH to the server and collect
some data from it. Tying the `add` command into Facter might be a nice addition
in the future.

Now we have two datacenters, we should reflect that. We'll pretend these are in
Amazon.

    for i in {101..120} ; do arsimto ln SF server-$i.sf ; done
    for i in {101..120} ; do arsimto ln OR server-$i.or ; done
    arsimto ln AWS OR
    arsimto ln AWS SF

Now let's see how things look:

    arsimto ls
	(AWS) --> () () 
	(OR) --> ++++++++++++++++++++
	(SF) --> ++++++++++++++++++++
	(memcached) --> ++++++++++++++++++++
	(mysql) --> ++++++++++
	(www) --> ++++++++++
	 - 6 pools.

If a pool points to another pool, it shows up as `(PointingPool) --> ()` and any
further assets are printed as `+`. To see the actual Asset or Pool, do `arsimto ls -l`.

Let's further put some Assets into production, and some into staging. Note that
Oregon and San Francisco aren't entirely symmetric here, adding some complexity
into our inventory.

	arsimto ln production server-{111..114}.sf
	arsimto ln production server-{111..114}.or
	arsimto ln staging server-{115..117}.sf
	arsimto ln staging server-{117..120}.or

Now let's move some servers around to complicate things further:

    arsimto mv Pools/memcached/server-11[4-6]* Pools/www/
    arsimto mv Pools/www/server-10[6-7]* Pools/mysql/

Now what do things look like?

	arsimto ls --intersect www OR --data=name,ip,disk,ram,network
	(OR+www)
	server-108.or	10.2.0.108	1tB	8GB	10gb
	server-109.or	10.2.0.109	1tB	8GB	10gb
	server-110.or	10.2.0.110	1tB	8GB	10gb
	server-114.or	10.2.0.114	8gB	32GB	40gb
	server-115.or	10.2.0.115	8gB	32GB	40gb
	server-116.or	10.2.0.116	8gB	32GB	40gb
	 - 6 assets in pool.
	 - 1 pools.

This is where an inventory management system starts to help matters. We've
added chaos into our system, and this helps us keep track of the chaos. Our WWW
pool is heterogenous, so we might weight the servers differently, or start
doubling up processes.

Here are a couple more queries to help us explore this inventory's topology:

	arsimto ls -l -i www production SF
	(SF+production+www) --> server-114.sf
	 - 1 pools.
	
	arsimto ls -l -i production memcached
	(memcached+production) --> server-111.or, server-111.sf, server-112.or, server-112.sf, server-113.or, server-113.sf
	 - 1 pools.

Quickly! How many production MySQL machines are there?

Now let's add a bit of hierarchy into our system. Both OR and SF are in USA:

	arsimto ln USA OR SF
	 - Connected USA --> OR
	 - Connected USA --> SF

By default arsimto will recurse one level into Pools looking for more pools to
list assets. So if you want to list all objects in the USA:

	arsimto ls USA -d=name,ip
	(OR)
	server-101.or	10.2.0.101
	...
	server-120.or	10.2.0.120
	 - 20 assets in pool.
	 (SF)
	server-101.sf	10.2.1.101
	...
	server-120.sf	10.2.1.120
	 - 20 assets in pool.
	 - 2 pools.

But suppose you put some asset directly into USA and don't want to recurse down
into SF and OR from USA:

	arsimto ls -l -R=0 USA -d=name,cages
	(USA)
	 + dc01	45
	 - 1 assets in pool.
	 - 1 pools.

Similarly, if you have a deeply-nested structure, you can specify -R=3 or beyond.
Be careful with this. If you have a lot of pools, this can quickly become a
performance killer. The recursion algorithm is naive.

Technical Notes
===============

You can do a lot of exploration outside the tool. Everything is stored as
directories and files in the CWD from where you run the tool. It's suggested to
keep these files in git or subversion so that users don't clobber each others'
changes, and you don't have to worry about permissions of shared directories.

The Unix filesystem is a tree structure. I am implementing a directed graph
structure on top of the Unix filesystem by storing Pool names and Asset names
in directories. The Assets/ directory could be thought of as a list of Nodes
with their related attributes. The Pools/ directory could be thought of as a
list of Left Nodes of Edges, and directories inside each are the Right Nodes of
the same Edges, with files inside the Right Node directory being the attributes
of the Edge.

The Unix filesystem has a very well-understood API. Because I'm not storing
data in the files, only file names, the API to traverse the data outside of
arsimto is "ls", "mv", "find", "rm", and "mkdir." You will likely be doing
"sort" and "grep" commands on the output reports occasionally.

FAQs
====

### Q: Reserved or Special Data key names?
A: Unfortunately, yes. `name`, `pools`, and any name starting with two
underscores (`__`, for example `__TYPE`) are reserved. Also `*` is reserved,
though you probably weren't thinking of using it.

The following are special names, and I recommend you use these precise key
names for the purpose outlined. In the future, the script might do extra
goodness based on them:

 * cpus :: Number of CPUs the system has.
 * diskMB :: `MB` of all disks combined on the Asset
 * [DEV]-diskMB :: `MB` of disk space available on device `DEV`.
 * dnsname :: The DNS name of the Asset. I use this as "desired DNS name" until
   it's actually in DNS, at which point it is "actual DNS name."
 * ip :: globally-addressable or public IP address. Any machine in your
   infrastructure that should be able to reach this Asset would reach
   it via this IP address. This is typically 192.168.0.0/16.
 * intip :: private IP address. Only certain machines would be able to
   reach this Asset via this IP address. This is typically 10.0.0.0/8.
 * mac :: The MAC address of the zeroeth NIC.
 * [IFACE]-mac :: The MAC address of the NIC mapped to `IFACE`.
 * mbit :: Network bandwidth of all interfaces on Asset combined.
 * [IFACE]-mbit :: Network bandwidth capability of `IFACE` in MBit/sec.
 * memkB :: kilobytes of RAM this Asset has. The `kB` capitalized like that
   is a consequence of how Linux reports memory in /proc/meminfo. I personally
   would prefer `KB`.

### Q: Backups?
A: Try tar -zcvPf backup.tgz /path/to/arsimto/AssetsPoolsDir. If you're
diligent about committing changes into git everytime you make them, then you
can also simply check out a previous revision.

### Q: Many simultaneous users?
A: Try putting Assets/ and Pools/ directories into git. This might also be
considered your "backup." If you're unfamiliar with git, have someone help you
set it up, then what you'll be doing every time you change your infrastructure
is:

	git pull
	git add Assets/ Pools/
	git commit -am 'added X servers, removed Y, changed RAM for Z'
	git push

### Q: Oh noes I deleted half my infrastructure!
A: `git reset --hard` will revert everything back to your previous commit/pull.

### Q: How can I tie this into Ansible?
A: You can write a plugin, but I'm too lazy. There's an example above about how
to do this, which I'll flesh out a little more here:

	ansibleInventory=~/tmp/ansibleInventory/inventoryFile.cfg
	echo "[RunOnTheseHosts]" > $ansibleInventory
	arsimto ls -p -i Pools/Production/ Pools/MySQLs/ -d=ip,memkB,dnsname \
		| awk '{print $1"   mem="$2"   dnsname="$3}' >> $ansibleInventory
	ansible-playbook -i $ansibleInventory playbook.yml

### Q: Pool nesting hierarchies?
A: No. Pool namespace is flat. See this example:

	arsimto ls -l Pools/
	(AWS) --> (OR) (SF)
	(OR)
	(Rackspace) --> (OR)
	(SF)
	(memcached)
	(mysql)
	(www)

AWS points to the OR (and SF) pool, but so does Rackspace. This doesn't mean
there are distinct AWS/OR and Rackspace/OR pools. It means they both point to
the same thing. This is almost certainly an error.

Because links are implemented as directories within a directory, you can fool
yourself by doing "arsimto ls Pools/GroupingPool/GroupedPool" and it will give
some sort of output (albeit confusing and "wrong").  This helps reinforce the
incorrect perception that there is nesting (aka Parent/Child) but such nesting
does not exist. I apologize for this confusing aspect now.

### Q: Cycles in the graph?
A: Sure. This makes sense for things like circular replication rings in MySQL,
for example. Here's what it looks like:

	arsimto ln server1234 server5678
	arsimto ln server5678 server8901
	arsimto ln server8901 server1234
	arsimto ls Pools/
	(server1234) --> (server5678)
	(server5678) --> (server8901)
	(server8901) --> (server1234)

### Q: Deleting a link between Pool and linked Pool/Asset?
A: This is not yet implemented, but true to Arsimto's simple form, you can do
it with a simple Unix command: `rm -rf Pools/Linker/Linkee` will remove the link
between the Linker and Linkee. This works whether the Linkee is another Pool or
an Asset.

Performance
===========

I decided to build a "big" inventory and do some timings. This initial setup
takes about five minutes on an SSD-based system. The script follows:

	arsimto add dc01   --data=capacity:5000
	arsimto add rack01   --data=U:48
	
	# shedload of servers
	arsimto add server{1000..2999} --data=ram:16GB,disk:2048GB,nic:10Gb
	arsimto add server{3000..5999} --data=ram:24GB,disk:4096GB,nic:1Gb+1Gb+1Gb+1Gb
	arsimto add server{6000..8999} --data=ram:32GB,disk:1024GB,nic:10Gb+10Gb
	
	# put servers in racks
	for i in {10..89} ; do for j in `seq -w 0 99` ; do arsimto ln rack$i server$i$j ; done ; done
	
	# put racks in cages
	for i in `seq -w 1 8` ; do for j in `seq -w 1 9` ; do arsimto ln cage$i rack$i$j ; done ; done
	
	# put cages in DC
	arsimto ln dc01 cage{1..8}
	
	# create some pools
	arsimto ln www server{10..89}0{1..9}
	arsimto ln www server{10..89}1{1..9}
	
	arsimto ln app server{10..89}2{1..9}
	arsimto ln app server{10..89}3{1..9}
	
	arsimto ln mysql server{10..89}4{1..9}
	arsimto ln mysql server{10..89}5{1..9}
	
	arsimto ln memcached server{10..89}6{1..9}
	arsimto ln memcached server{10..89}7{1..9}
	arsimto ln memcached server{10..89}8{1..9}
	
	arsimto ln varnish server{10..89}9{1..9}

	arsimto ln production server{10..89}{0..9}{0..5}
	arsimto ln dev        server{10..89}{0..9}{6..7}
	arsimto ln staging    server{10..89}{0..9}{8..9}
	
	# add joke MAC for some servers
	for i in `seq -w 10 89` ; do for j in `seq 1 9` ; do arsimto add server${i}1${j} --data=mac:aa:bb:$i:00:f$j:dd:e0:f1 ; done ; done
	for i in `seq -w 10 89` ; do for j in `seq 1 9` ; do arsimto add server${i}3${j} --data=mac:bb:cc:$i:00:f$j:dd:e0:f1 ; done ; done
	for i in `seq -w 10 89` ; do for j in `seq 1 9` ; do arsimto add server${i}5${j} --data=mac:cc:dd:$i:00:f$j:dd:e0:f1 ; done ; done
	for i in `seq -w 10 89` ; do for j in `seq 1 9` ; do arsimto add server${i}7${j} --data=mac:dd:ee:$i:00:f$j:dd:e0:f1 ; done ; done

Now let's do some timings!

	for i in app varnish memcached mysql ; do echo $i ----------- ; time arsimto ls $i >/dev/null ; done
	app -----------
	real	0m0.071s
	user	0m0.047s
	sys	0m0.023s
	varnish -----------
	real	0m0.037s
	user	0m0.024s
	sys	0m0.012s
	memcached -----------
	real	0m0.098s
	user	0m0.064s
	sys	0m0.033s
	mysql -----------
	real	0m0.068s
	user	0m0.045s
	sys	0m0.022s
	
	time arsimto ls mysql --data=ram,disk,mac >/dev/null
	real	0m0.079s
	user	0m0.055s
	sys	0m0.024s

	time arsimto ls -l memcached --data=name,ram,disk,mac >/dev/null
	real	0m0.114s
	user	0m0.079s
	sys	0m0.034s

	time arsimto ls -i production memcached -d=name,ram,mac >/dev/null
	real	0m0.308s
	user	0m0.208s
	sys	0m0.098s
	
	time arsimto ls -i production memcached rack44 -d=name,ram,mac >/dev/null
	real	0m0.306s
	user	0m0.206s
	sys	0m0.098s
	
	time arsimto rename server8696 varnish8696
	 - Renamed server8696 --> varnish8696
	 real	0m0.523s
	 user	0m0.063s
	 sys	0m0.440s
	
	time arsimto rm server-34{2..5}{2..5} >/dev/null #(16 servers)
	real	0m5.358s
	user	0m1.106s
	sys	0m4.244s

Note that these timings were largely coming from cache, so they represent best-case timings.
If you are hitting SSDs, they are about 10x slower, and spinning disk should be about 100x
slower.

