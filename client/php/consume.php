<?php
require './vendor/autoload.php';
date_default_timezone_set('PRC');

use Monolog\Logger;
use Monolog\Handler\StreamHandler; 

// Create the logger
$logger = new Logger('my_logger');
// Now add some handlers
$logger->pushHandler(new StreamHandler('test.log', Logger::WARNING));

$config = \Kafka\ConsumerConfig::getInstance();
$config->setMetadataRefreshIntervalMs(10000);
$config->setMetadataBrokerList('192.168.0.11:9092');
$config->setGroupId('mytest-1');
$config->setBrokerVersion('1.0.0');
$config->setTopics(['test']);
//$config->setOffsetReset('earliest');
$consumer = new \Kafka\Consumer();
$consumer->setLogger($logger);
$consumer->start(function($topic, $part, $message) {
    echo "topic ".$topic." part: ".$part.PHP_EOL;
    var_dump($message);
    echo PHP_EOL;
});

/**
 * 发送消息
 * php producter.php
 * 
 * 消费消息
 * php consume.php 启动一个守护job进行消费message
 * 输出结果：
topic test part: 0
array(3) {
  ["offset"]=>
  int(40299)
  ["size"]=>
  int(56)
  ["message"]=>
  array(6) {
    ["crc"]=>
    int(3893961150)
    ["magic"]=>
    int(1)
    ["attr"]=>
    int(0)
    ["timestamp"]=>
    int(-1)
    ["key"]=>
    string(10) "mytest:php"
    ["value"]=>
    string(24) "hello,current index:9998"
  }
}

topic test part: 0
array(3) {
  ["offset"]=>
  int(40300)
  ["size"]=>
  int(56)
  ["message"]=>
  array(6) {
    ["crc"]=>
    int(2669547816)
    ["magic"]=>
    int(1)
    ["attr"]=>
    int(0)
    ["timestamp"]=>
    int(-1)
    ["key"]=>
    string(10) "mytest:php"
    ["value"]=>
    string(24) "hello,current index:9999"
  }
}
 */