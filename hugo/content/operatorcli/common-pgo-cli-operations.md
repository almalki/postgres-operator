---
title: "Common Operations"
date:
draft: false
weight: 20
---

## Common Operations

In all the examples below, the user is specifying the *pgouser1* namespace
as the target of the operator.  Replace this value with your own namespace
value.  You can specify a default namespace to be used by setting the
PGO_NAMESPACE environment variable on the `pgo` client environment.

### Cluster Operations

A user will typically start using the Operator by creating a Postgres
cluster as follows:

    pgo create cluster mycluster -n pgouser1

This command creates a Postgres cluster in the *pgouser1* namespace
that has a single Postgres primary container.

You can see the Postgres cluster using the following:

    pgo show cluster mycluster -n pgouser1

You can test the Postgres cluster by entering:

    pgo test mycluster -n pgouser1

You can optionally add a Postgres replica to your Postgres
cluster as follows:

    pgo scale mycluster -n pgouser1

You can create a Postgres cluster initially with a Postgres replica as follows:

    pgo create cluster mycluster --replica-count=1 -n pgouser1

To view the Postgres logs, you can enter commands such as:

    pgo ls mycluster -n pgouser1 /pgdata/mycluster/pg_log
    pgo cat mycluster -n pgouser1 /pgdata/mycluster/pg_log/postgresql-Mon.log | tail -3


#### Backups

By default the Operator deploys pgbackrest for a Postgres cluster to
hold database backup data.

You can create a pgbackrest backup job as follows:

    pgo backup mycluster -n pgouser1

You can perform a pgbasebackup job as follows:

    pgo backup mycluster --backup-type=pgbasebackup -n pgouser1

You can optionally pass pgbackrest command options into the backup
command as follows:

    pgo backup mycluster --backup-type=pgbackrest --backup-opts="--type=diff" -n pgouser1

See pgbackrest.org for command flag descriptions.

You can create a Postgres cluster that does not include pgbackrest
if you specify the following:

    pgo create cluster mycluster --pgbackrest=false -n pgouser1

You can show the current backups on a cluster with the following:

    pgo show backup mycluster -n pgouser1

#### Scaledown a Cluster

You can remove a Postgres replica using the following:

    pgo scaledown mycluster --query -n pgouser1
    pgo scaledown mycluster --target=sometarget -n pgouser1

#### Delete a Cluster

You can remove a PostgreSQL cluster by entering:

    pgo delete cluster mycluster -n pgouser1

This removes any PostgreSQL instances from being accessed as well as deletes all of its data and backups.

##### Retain Backups

It can often be useful to keep the backups of a cluster even after its deleted, such as for archival purposes or for creating the cluster at a future date. You can delete the cluster but keep its backups using the `--keep-backups` flag:

```bash
pgo delete cluster mycluster --keep-backups -n pgouser1
```

##### Retain Cluster Data

There are rare circumstances in which you may want to keep a copy of the original cluster data, such as when upgrading manually to a newer version of the Operator. In these cases, you can use the `--keep-data` flag:

```bash
pgo delete cluster mycluster --keep-data -n pgouser1
```

**NOTE**: The `--keep-data` flag is deprecated.

#### View Disk Utilization

You can see a comparison of Postgres data size versus the Persistent
volume claim size by entering the following:

    pgo df mycluster -n pgouser1

### Label Operations
#### Apply a Label to a Cluster

You can apply a Kubernetes label to a Postgres cluster as follows:

    pgo label mycluster --label=environment=prod -n pgouser1

In this example, the label key is *environment* and the label
value is *prod*.

You can apply labels across a category of Postgres clusters by
using the *--selector* command flag as follows:

    pgo label --selector=clustertypes=research --label=environment=prod -n pgouser1

In this example, any Postgres cluster with the label of *clustertypes=research*
will have the label *environment=prod* set.

In the following command, you can also view Postgres clusters by
using the *--selector* command flag which specifies a label key value
to search with:

    pgo show cluster --selector=environment=prod -n pgouser1

### Policy Operations
#### Create a Policy

To create a SQL policy, enter the following:

    pgo create policy mypolicy --in-file=mypolicy.sql -n pgouser1

This examples creates a policy named *mypolicy* using the contents
of the file *mypolicy.sql* which is assumed to be in the current
directory.

You can view policies as following:

    pgo show policy --all -n pgouser1


#### Apply a Policy

    pgo apply mypolicy --selector=environment=prod
    pgo apply mypolicy --selector=name=mycluster

### Operator Status
#### Show Operator Version

To see what version of the Operator client and server you are using, enter:

    pgo version

To see the Operator server status, enter:

    pgo status -n pgouser1

To see the Operator server configuration, enter:

    pgo show config -n pgouser1

To see what namespaces exist and if you have access to them, enter:

    pgo show namespace pgouser1

#### Perform a pgdump backup

    pgo backup mycluster --backup-type=pgdump -n pgouser1
    pgo backup mycluster --backup-type=pgdump --backup-opts="--dump-all --verbose" -n pgouser1
    pgo backup mycluster --backup-type=pgdump --backup-opts="--schema=myschema" -n pgouser1

Note: To run `pgdump_all` instead of `pgdump`, pass `--dump-all` flag in `--backup-opts` as shown above. All `--backup-opts` should be space delimited.

#### Perform a pgbackrest restore

    pgo restore mycluster -n pgouser1

Or perform a restore based on a point in time:

    pgo restore mycluster --pitr-target="2019-01-14 00:02:14.921404+00" --backup-opts="--type=time" -n pgouser1

You can also set the any of the [pgbackrest restore options](https://pgbackrest.org/command.html#command-restore) :

    pgo restore mycluster --pitr-target="2019-01-14 00:02:14.921404+00" --backup-opts=" see pgbackrest options " -n pgouser1

You can also target specific nodes when performing a restore:

    pgo restore mycluster --node-label=failure-domain.beta.kubernetes.io/zone=us-central1-a -n pgouser1

Here are some steps to test PITR:

 * `pgo create cluster mycluster`
 * Create a table on the new cluster called *beforebackup*
 * pgo backup mycluster -n pgouser1
 * create a table on the cluster called *afterbackup*
 * Execute *select now()* on the database to get the time, use this timestamp minus a couple of minutes when you perform the restore
 * `pgo restore mycluster --pitr-target="2019-01-14 00:02:14.921404+00" --backup-opts="--type=time --log-level-console=info" -n pgouser1`
 * Wait for the database to be restored
 * Execute *\d* in the database and you should see the database state prior to where the *afterbackup* table was created

See the Design section of the Operator documentation for things to consider
before you do a restore.

#### Restore from pgbasebackup

You can find available pgbasebackup backups to use for a pgbasebackup restore using the `pgo show backup` command:

```
$ pgo show backup mycluster --backup-type=pgbasebackup -n pgouser1 | grep "Backup Path"
        Backup Path:    mycluster-backups/2019-05-21-09-53-20
        Backup Path:    mycluster-backups/2019-05-21-06-58-50
        Backup Path:    mycluster-backups/2019-05-21-09-52-52
```

You can then perform a restore using any available backup path:

    pgo restore mycluster --backup-type=pgbasebackup --backup-path=mycluster/2019-05-21-06-58-50 --backup-pvc=mycluster-backup -n pgouser1

When performing the restore, both the backup path and backup PVC can be omitted, and the Operator will use the last pgbasebackup backup created, along with the PVC utilized for that backup:

    pgo restore mycluster --backup-type=pgbasebackup -n pgouser1

Once the pgbasebackup restore is complete, a new PVC will be available with a randomly generated ID that contains the restored database, e.g. PVC  **mycluster-ieqe** in the output below:

```
$ pgo show pvc --all
All Operator Labeled PVCs
        mycluster
        mycluster-backup
        mycluster-ieqe
```

A new cluster can then be created with the same name as the new PVC, as well with the secrets from the original cluster, in order to deploy a new cluster using the restored database:

    pgo create cluster mycluster-ieqe --secret-from=mycluster

If you would like to control the name of the PVC created when performing a pgbasebackup restore, use the `--restore-to-pvc` flag:

    pgo restore mycluster --backup-type=pgbasebackup --restore-to-pvc=mycluster-restored -n pgouser1

#### Restore from pgdump backup

    pgo restore mycluster --backup-type=pgdump --backup-pvc=mycluster-pgdump-pvc --pitr-target="2019-01-15-00-03-25" -n pgouser1

To restore the most recent pgdump at the default path, leave off a timestamp:

    pgo restore mycluster --backup-type=pgdump --backup-pvc=mycluster-pgdump-pvc -n pgouser1


### Fail-over Operations

To perform a manual failover, enter the following:

    pgo failover mycluster --query -n pgouser1

That example queries to find the available Postgres replicas that
could be promoted to the primary.

    pgo failover mycluster --target=sometarget -n pgouser1

That command chooses a specific target, and starts the failover workflow.

#### Create a Cluster with Auto-fail Enabled

To support an automated failover, you can specify the *--autofail* flag
on a Postgres cluster when you create it as follows:

    pgo create cluster mycluster --autofail=true --replica-count=1 -n pgouser1

You can set the auto-fail flag on a Postgres cluster after it is created
by the following command:

    pgo update cluster --autofail=false -n pgouser1
    pgo update cluster --autofail=true -n pgouser1

Note that if you do a pgbackrest restore, you will need to reset the
autofail flag to true after the restore is completed.

### Add-On Operations

To add a pgbouncer Deployment to your Postgres cluster, enter:

    pgo create cluster mycluster --pgbouncer -n pgouser1

You can add pgbouncer after a Postgres cluster is created as follows:

    pgo create pgbouncer mycluster
    pgo create pgbouncer --selector=name=mycluster

You can also specify a pgbouncer password as follows:

    pgo create cluster mycluster --pgbouncer --pgbouncer-pass=somepass -n pgouser1

Note, the pgbouncer configuration defaults to specifying only
a single entry for the primary database.  If you want it to
have an entry for the replica service, add the following
configuration to pgbouncer.ini:

    {{.PG_REPLICA_SERVICE_NAME}} = host={{.PG_REPLICA_SERVICE_NAME}} port={{.PG_PORT}} auth_user={{.PG_USERNAME}} dbname={{.PG_DATABASE}}

You can remove a pgbouncer from a cluster as follows:

    pgo delete pgbouncer mycluster -n pgouser1

You can create a pgbadger sidecar container in your Postgres cluster
pod as follows:

    pgo create cluster mycluster --pgbadger -n pgouser1

Likewise, you can add the Crunchy Collect Metrics sidecar container
into your Postgres cluster pod as follows:

    pgo create cluster mycluster --metrics -n pgouser1

Note: backend metric storage such as Prometheus and front end
visualization software such as Grafana are not created automatically
by the PostgreSQL Operator.  For instructions on installing Grafana and
Prometheus in your environment, see the [Crunchy Container Suite documentation](https://access.crunchydata.com/documentation/crunchy-containers/4.2.0/examples/metrics/metrics/).

### Scheduled Tasks

There is a cron based scheduler included into the Operator Deployment
by default.

You can create automated full pgBackRest backups every Sunday at 1 am
as follows:

    pgo create schedule mycluster --schedule="0 1 * * SUN" \
        --schedule-type=pgbackrest --pgbackrest-backup-type=full -n pgouser1

You can create automated diff pgBackRest backups every Monday-Saturday at 1 am
as follows:

    pgo create schedule mycluster --schedule="0 1 * * MON-SAT" \
        --schedule-type=pgbackrest --pgbackrest-backup-type=diff -n pgouser1

You can create automated pgBaseBackup backups every day at 1 am as
follows:

In order to have a backup PVC created, users should run the `pgo backup` command
against the target cluster prior to creating this schedule.

    pgo create schedule mycluster --schedule="0 1 * * *" \
        --schedule-type=pgbasebackup --pvc-name=mycluster-backup -n pgouser1

You can create automated Policy every day at 1 am as follows:

    pgo create schedule --selector=pg-cluster=mycluster --schedule="0 1 * * *" \
         --schedule-type=policy --policy=mypolicy --database=userdb \
         --secret=mycluster-testuser-secret -n pgouser1

### Benchmark Clusters

The pgbench utility containerized and made available to Operator
users.

To create a Benchmark via Cluster Name you enter:

    pgo benchmark mycluster -n pgouser1

To create a Benchmark via Selector, enter:

    pgo benchmark --selector=pg-cluster=mycluster -n pgouser1

To create a Benchmark with a custom transactions, enter:

    pgo create policy --in-file=/tmp/transactions.sql mytransactions -n pgouser1
    pgo benchmark mycluster --policy=mytransactions -n pgouser1

To create a Benchmark with custom parameters, enter:

    pgo benchmark mycluster --clients=10 --jobs=2 --scale=10 --transactions=100 -n pgouser1

You can view benchmarks by entering:

    pgo show benchmark -n pgouser1 mycluster

### Complex Deployments
#### Create a Cluster using Specific Storage

    pgo create cluster mycluster --storage-config=somestorageconfig -n pgouser1

Likewise, you can specify a storage configuration when creating
a replica:

    pgo scale mycluster --storage-config=someslowerstorage -n pgouser1

This example specifies the *somestorageconfig* storage configuration
to be used by the Postgres cluster.  This lets you specify a storage
configuration that is defined in the *pgo.yaml* file specifically for
a given Postgres cluster.

You can create a Cluster using a Preferred Node as follows:

    pgo create cluster mycluster --node-label=speed=superfast -n pgouser1

That command will cause a node affinity rule to be added to the
Postgres pod which will influence the node upon which Kubernetes
will schedule the Pod.

Likewise, you can create a Replica using a Preferred Node as follows:

    pgo scale mycluster --node-label=speed=slowerthannormal -n pgouser1

#### Create a Cluster with LoadBalancer ServiceType

    pgo create cluster mycluster --service-type=LoadBalancer -n pgouser1

This command will cause the Postgres Service to be of a specific
type instead of the default ClusterIP service type.

#### Namespace Operations

Create an Operator namespace where Postgres clusters can be created
and managed by the Operator:

    pgo create namespace mynamespace

Update a Namespace to be able to be used by the Operator:

    pgo update namespace somenamespace

Delete a Namespace:

    pgo delete namespace mynamespace

#### PGO User Operations

PGO users are users defined for authenticating to the PGO REST API.  You
can manage those users with the following commands:

    pgo create pgouser someuser --pgouser-namespaces="pgouser1,pgouser2" --pgouser-password="somepassword" --pgouser-roles="pgoadmin"
    pgo create pgouser otheruser --all-namespaces --pgouser-password="somepassword" --pgouser-roles="pgoadmin"

Update a user:

    pgo update pgouser someuser --pgouser-namespaces="pgouser1,pgouser2" --pgouser-password="somepassword" --pgouser-roles="pgoadmin"
    pgo update pgouser otheruser --all-namespaces --pgouser-password="somepassword" --pgouser-roles="pgoadmin"

Delete a PGO user:

    pgo delete pgouser someuser

PGO roles are also managed as follows:

    pgo create pgorole somerole --permissions="Cat,Ls"

Delete a PGO role with:

    pgo delete pgorole somerole

Update a PGO role with:

    pgo update pgorole somerole --permissions="Cat,Ls"

#### Postgres User Operations

Managed Postgres users can be viewed using the following command:

    pgo show user mycluster

Postgres users can be created using the following command examples:

    pgo create user mycluster --username=somepguser --password=somepassword --managed
    pgo create user --selector=name=mycluster --username=somepguser --password=somepassword --managed

Those commands are identical in function, and create on the mycluster Postgres cluster, a user named *somepguser*, with a password of *somepassword*, the account is *managed* meaning that
these credentials are stored as a Secret on the Kubernetes cluster in the Operator
namespace.

Postgres users can be deleted using the following command:

    pgo delete user mycluster --username=somepguser

That command deletes the user on the mycluster Postgres cluster.

Postgres users can be updated using the following command:

    pgo update user mycluster --username=somepguser --password=frodo

That command changes the password for the user on the mycluster Postgres cluster.


#### Miscellaneous

Create a cluster using the Crunchy Postgres + PostGIS container image:

    pgo create cluster mygiscluster --ccp-image=crunchy-postgres-gis -n pgouser1

Create a cluster with a Custom ConfigMap:

    pgo create cluster mycustomcluster --custom-config myconfigmap -n pgouser1
