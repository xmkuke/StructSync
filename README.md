# Struct Sync
Mysql table structure automatic synchronization tool

Used to synchronize the `online` database structure <b>change</b> to `local environment`!
Support function:
1. Sync ** new table **
2. Sync** field ** Change: Add, modify, delete
3. Sync ** Index ** Change: Add, modify, delete
4. Support **Preview** (compares struct and save to file, not execute)
5. Support local extra lines, additional tables, fields, indexes, foreign keys


### Installation
>go get -u github.com/xmkuke/StructSync


### Configuration
Reference Default configuration file config.json Configure the synchronization source and destination address.
Modifying the recipient of the email You can receive an email notification when the operation fails or there is a change in the table structure.

By default, no extra **lists, fields, indexes, foreign keys** are deleted. If you need to delete the ** field, index, foreign key ** you can use the <code>-drop</code> parameter.

Configuration example (app.conf):
```javascript
{
  "SrcDbDsn": {
    "Host": "127.0.0.1",
    "Port": "3306",
    "DbName": "test_std",
    "User": "root",
    "Pswd": "",
    "Charset": "utf8"
  },
  "DestDbList": [{
    "Host": "127.0.0.1",
    "Port": "3306",
    "DbName": "test_1",
    "User": "root",
    "Pswd": "Aa123654",
    "Charset": "utf8"
  },
    {
      "Host": "127.0.0.1",
      "Port": "3306",
      "DbName": "test_2",
      "User": "root",
      "Pswd": "",
      "Charset": "utf8"
    }
  ],
  "PageSize": 20,
  "ChanNum": 4,
  "InputSql": "",
  "OutputDir": "./output",
  "DropUnecessary": false,
  "InputMode": 1,
  "ExecuteSQL": false,
  "SaveSQL": true,
  "TimeOut": "600s",
  "LogLevel": 2,
  "LogPath": "",
  "LogFileName": "StructSync_${date}.log"
}
```

#### json configuration item description
SrcDbDsn: database synchronization source
DestDbList: database to be synchronized, use array specify multiple databases
ChanNum: Specify how many coroutines to execute simultaneously
OutputDir: Save the adjusted SQL directory 
DropUnecessary: Whether to delete extra fields or indexes, not delete by default
InputMode: 1 Use standard database, 2 use schema file (you can export a database schema to file)
ExecuteSQL: Whether to automatically perform the adjusted SQL to the target database, the default is to execute
SaveSQL: Whether to save the adjusted SQL to the file
TimeOut: Execute SQL timeout, default 600s(The length of time to adjust the database structure will vary depending on the amount of data in the database itself.)
LogLevel: Display the log level of the execution record, ALL-0，DEBUG-1，INFO-2，WARN-3，ERROR-4，FATAL-5，OFF-6 
LogPath: Log path
LogFileName: Log filename, can use ${data} or ${time} param, default is 'StructSync_${date}.log'

### Running
### Param & Usage
```
Usage of ./StructSync:
  -c    Use the param execute delete unnecessary field / index 
  -e    Execute adjust SQL to dest database, default true (default true)
  -i string
        Default read source schema info from database， use -i，read source schema info from file
  -o string
        Save adjust SQL to file

```

### Direct operation
```
./StructSync 
```
 
### Show Help
```
./StructSync --help
```
### Save SQL and not execute
```
./StructSync -e false -o "./output/adjust.sql"
```

Each json file is configured with a destination database, and the check.sh script runs each configuration in turn.
The log is stored in the current log directory.

### Automatic timing operation
Add crontab task

<code>
30 * * * * cd /your/path/xxx/ && bash check.sh >/dev/null 2>&1
</code>

### Parameter Description
<code>
./StructSync --help
</code>

Description:
<pre><code>
# ./StructSync --help
Usage of ./StructSync:
  -c    Use the param execute delete unnecessary field / index 
  -e    Execute adjust SQL to dest database, default true (default true)
  -i string
        Default read source schema info from database， use -i，read source schema info from file
  -o string
        Save adjust SQL to file

</code>
</pre>