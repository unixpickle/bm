# bm - **b**ook**m**ark

This is a simple tool to bookmark commands that you enter on the command-line and search them by name or content.

# Usage

## Named bookmarks

Let's say you want to run a command and you want to bookmark it so you can find it again later. The simplest thing to do is create a named bookmark, using `bookmark cn`, like so:

```
$ bm cn bookmarkname ls -l /
created record with ID bookmarkname
running command 'bookmarkname': ls -l /
total 9
drwxrwxr-x+ 17 root  admin   544 Sep  3 14:25 Applications
drwxr-xr-x  69 root  wheel  2208 Aug 15 14:23 Library
drwxr-xr-x@  8 root  wheel   256 Apr  6 15:46 System
drwxr-xr-x   5 root  admin   160 Apr  6 15:45 Users
drwxr-xr-x   3 root  wheel    96 Sep 16 15:59 Volumes
...
```

In the above example, we made a bookmark called `bookmarkname` with the command `ls -l /`. Now if you ever want to run this command again, you can simply use `bm rn`:

```
$ bm rn bookmarkname
running command 'bookmarkname': ls -l /
total 9
drwxrwxr-x+ 17 root  admin   544 Sep  3 14:25 Applications
drwxr-xr-x  69 root  wheel  2208 Aug 15 14:23 Library
drwxr-xr-x@  8 root  wheel   256 Apr  6 15:46 System
drwxr-xr-x   5 root  admin   160 Apr  6 15:45 Users
drwxr-xr-x   3 root  wheel    96 Sep 16 15:59 Volumes
...
```

To see a list of all your bookmarks, use `bm a` like so:

```
$ bm a
 bookmarkname  ls -l /
```

To query for a specific bookmark name, use `bm qn`:

```
$ bm qn book
 bookmarkname  ls -l /
```

To delete a bookmark by name, use `bm dn bookmarkname`. Note that this command soft-matches the name if no exact matches are found, so you should use caution when deleting bookmarks.

## Unnamed bookmarks

For every command, like `cn`, `dn`, or `rn`, there is an equivalent command like `c`, `d`, and `r` that matches command contents instead of addressing commands by name. For example, if you wanted to find the bookmark we created above, but just remembered it had "-l" somewhere in it, you could use `bm q -l`:

```
$ bm q -l
 bookmarkname  ls -l /
```

When you use the `bm c` command, simply omit the bookmark name and a random one will be generated.

```
$ bm c echo hello world
created record with ID 0
running command '0': echo hello world
hello world
```

# Related projects

**bm** is similar to [Marker](https://github.com/pindexis/marker), but it is written in pure Go and therefore more portable.
