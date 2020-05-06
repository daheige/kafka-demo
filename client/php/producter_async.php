<?php
require './vendor/autoload.php';
date_default_timezone_set('PRC');

use Monolog\Logger;
use Monolog\Handler\StreamHandler; 

// Create the logger
$logger = new Logger('my_logger');
// Now add some handlers
$logger->pushHandler(new StreamHandler('test.log', Logger::WARNING));

/**
 * 异步发送消息模式
 */
$config = \Kafka\ProducerConfig::getInstance();
$config->setMetadataRefreshIntervalMs(10000);
$config->setMetadataBrokerList('192.168.0.11:9092');
$config->setBrokerVersion('1.0.0');
$config->setRequiredAck(1);
$config->setIsAsyn(false);
$config->setProduceInterval(500);
$producer = new \Kafka\Producer(
    function() {
        return [
            [
                'topic' => 'test',
                'value' => 'hello,daheige123',
                'key' => 'mytest:php',
            ],
        ];
    }
);
$producer->setLogger($logger);
$producer->success(function($result) {
	var_dump($result);
});
$producer->error(function($errorCode) {
		var_dump($errorCode);
});
$producer->send(true);