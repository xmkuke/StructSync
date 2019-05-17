数据库说明：
业务数据库需新建同步日志表

CREATE TABLE `com_db_sync_log` (
	`id` INT(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
	`sync_key` VARCHAR(32) NULL DEFAULT '' COMMENT '随机key',
	`db_id` INT(10) UNSIGNED NOT NULL DEFAULT '0' COMMENT '数据库ID',
	`ret` INT(10) UNSIGNED NULL DEFAULT '0' COMMENT '同步结果：0-全部失败，1-全部成功，2-部分失败',
	`add_time` DATETIME NULL DEFAULT NULL COMMENT '添加时间',
	`update_time` DATETIME NULL DEFAULT NULL COMMENT '更新时间',
	PRIMARY KEY (`id`),
	INDEX `sync_key_db_id` (`sync_key`, `db_id`)
)
COMMENT='数据库同步日志表'
COLLATE='utf8_general_ci'
ENGINE=InnoDB
ROW_FORMAT=COMPACT
;


命令参数说明：

Usage of mysql_ds_sync.exe:

  -c string：是否同步删除目标库多余字段:n-否,y-是 (default "n")

  -e string：SQL运行模式:r-直接运行,s-保存至文件,both-运行并保存 (default "r")

  -i string：SQL文件绝对路径,仅当m=input时有效

  -m string：启动模式:auto-数据库自动同步数据结构,input-通过导入SQL文件同步 (default "auto")

  -o string：SQL输出目录,仅当e!=r有效

脚本文件格式说明：

1、每条语句请使用";"结尾

2、执行多条语句，各语句请独立一行，工具不支持多查询


执行命令示例：

保存脚本： mysql_ds_sync.exe -c=y -e=s -o=E:/

执行并保存： mysql_ds_sync.exe -c=y -e=both -o=E:/

执行文件： mysql_ds_sync.exe -c=y -m=input -i=D:/t3.sql

