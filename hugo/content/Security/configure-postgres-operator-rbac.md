---
title: "Configuration of PostgreSQL Operator RBAC"
date:
draft: false
weight: 7
---


## PostreSQL Operator RBAC

The *conf/postgres-operator/pgorole* file is read at start up time when the operator is deployed to the Kubernetes cluster.  This file defines the PostgreSQL Operator roles whereby PostgreSQL Operator API users can be authorized.

The *conf/postgres-operator/pgouser* file is read at start up time also and contains username, password, role, and namespace information as follows:

    username:password:pgoadmin:
    pgouser1:password:pgoadmin:pgouser1
    pgouser2:password:pgoadmin:pgouser2
    pgouser3:password:pgoadmin:pgouser1,pgouser2
    readonlyuser:password:pgoreader:

The format of the pgouser server file is:

    <username>:<password>:<role>:<namespace,namespace>

The namespace is a comma separated list of namespaces that user has access to.  If you do not specify a namespace, then all namespaces is assumed, meaning this user can access any namespace that the Operator is watching.

A user creates a *.pgouser* file in their $HOME directory to identify themselves to the Operator.  An entry in .pgouser will need to match entries in the *conf/postgres-operator/pgouser* file.  A sample *.pgouser* file contains the following:

    username:password

The format of the .pgouser client file is:

    <username>:<password>

The users pgouser file can also be located at:

*/etc/pgo/pgouser* 

or it can be found at a path specified by the PGOUSER environment variable.

If the user tries to access a namespace that they are not configured for within the server side *pgouser* file then they will get an error message as follows:

    Error: user [pgouser1] is not allowed access to namespace [pgouser2]


The following list shows the current complete list of possible pgo permissions that you can specify within the *pgorole* file when creating roles:

|Permission|Description  |
|---|---|
|ApplyPolicy | allow *pgo apply*|
|Cat | allow *pgo cat*|
|CreateBackup | allow *pgo backup*|
|CreateBenchmark | allow *pgo create benchmark*|
|CreateCluster | allow *pgo create cluster*|
|CreateDump | allow *pgo create pgdump*|
|CreateFailover | allow *pgo failover*|
|CreatePgbouncer | allow *pgo create pgbouncer*|
|CreatePgpool | allow *pgo create pgpool*|
|CreatePolicy | allow *pgo create policy*|
|CreateSchedule | allow *pgo create schedule*|
|CreateUpgrade | allow *pgo upgrade*|
|CreateUser | allow *pgo create user*|
|DeleteBackup | allow *pgo delete backup*|
|DeleteBenchmark | allow *pgo delete benchmark*|
|DeleteCluster | allow *pgo delete cluster*|
|DeletePgbouncer | allow *pgo delete pgbouncer*|
|DeletePgpool | allow *pgo delete pgpool*|
|DeletePolicy | allow *pgo delete policy*|
|DeleteSchedule | allow *pgo delete schedule*|
|DeleteUpgrade | allow *pgo delete upgrade*|
|DeleteUser | allow *pgo delete user*|
|DfCluster | allow *pgo df*|
|Label | allow *pgo label*|
|Load | allow *pgo load*|
|Ls | allow *pgo ls*|
|Reload | allow *pgo reload*|
|Restore | allow *pgo restore*|
|RestoreDump | allow *pgo restore* for pgdumps|
|ShowBackup | allow *pgo show backup*|
|ShowBenchmark | allow *pgo show benchmark*|
|ShowCluster | allow *pgo show cluster*|
|ShowConfig | allow *pgo show config*|
|ShowPolicy | allow *pgo show policy*|
|ShowPVC | allow *pgo show pvc*|
|ShowSchedule | allow *pgo show schedule*|
|ShowNamespace | allow *pgo show namespace*|
|ShowUpgrade | allow *pgo show upgrade*|
|ShowWorkflow | allow *pgo show workflow*|
|Status | allow *pgo status*|
|TestCluster | allow *pgo test*|
|UpdateCluster | allow *pgo update cluster*|
|User | allow *pgo user*|
|Version | allow *pgo version*|


If the user is unauthorized for a pgo command, the user will get back this response:

    Error:  Authentication Failed: 401 

## Making Security Changes

Importantly, it is necesssary to redeploy the PostgreSQL Operator prior to giving effect to the user security changes in the pgouser and pgorole files:

    make deployoperator

Performing this command will recreate the *pgo-config* ConfigMap that stores these files and is mounted by the Operator during its initialization.