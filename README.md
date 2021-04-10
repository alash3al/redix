Redix v2
========
> here is the home of v2 which is under construction

Thoughts
=========
- Enterprise ready
- New Modules System to allow premium commands, features.
- There should be a premium middleware that changes that tells the context which db (namespace) should be used, tell now I'm thinking about creating a central databases table that hold the database name (i.e "0"), internal id and token_id (who owns the db), then the middleware should detect the actual db via the token + dbname, so the client must use AUTH before anything.
- If we removed that middleware, the context will use the db directly without any hassle. 