<?php
require './vendor/autoload.php';
date_default_timezone_set('PRC');

use Monolog\Logger;
use Monolog\Handler\StreamHandler; 

// Create the logger
$logger = new Logger('my_logger');
// Now add some handlers
$logger->pushHandler(new StreamHandler('./test.log', Logger::WARNING));

$logger->error('aa',["aa"=>1,'b'=>'123']);

$config = \Kafka\ProducerConfig::getInstance();
$config->setMetadataRefreshIntervalMs(10000);
$config->setMetadataBrokerList('192.168.0.11:9092');
$config->setBrokerVersion('1.0.0');
$config->setRequiredAck(1);
$config->setIsAsyn(false);
$config->setProduceInterval(500);
$producer = new \Kafka\Producer();
$producer->setLogger($logger);

for($i = 0; $i < 10000; $i++) {
    $producer->send([
        [
            'topic' => 'test',
            'value' => 'hello,current index:'.$i,
            'key' => 'mytest:php',
        ],
    ]);
}

/**
 * 运行同步发送消息
 * php producter.php
 * 启动go consume进行消费
 * go run consume/consume.go
 * 输出结果如下：
   message at offset 10295: mytest:php = hello,current index:9994
   current group_id:  mytest-1
   message at offset 10296: mytest:php = hello,current index:9995
   current group_id:  mytest-1
   message at offset 10297: mytest:php = hello,current index:9996
   current group_id:  mytest-1
   message at offset 10298: mytest:php = hello,current index:9997
   current group_id:  mytest-1
   message at offset 10299: mytest:php = hello,current index:9998
   current group_id:  mytest-1
   message at offset 10300: mytest:php = hello,current index:9999
 */
