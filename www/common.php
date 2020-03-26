<?php
function push($s, $HOST, $UID) {
	require __DIR__ . '/vendor/autoload.php';
	Predis\Autoloader::register();

	$client = new Predis\Client([
		'scheme' => 'tcp',
		'host'   => 'localhost',
		'port'   => 6379,
	]);

	if (isset($_SESSION['ip']))
		$client->del($_SESSION['ip']);

	if ($HOST != "") {
		$client->hset($HOST, 'd', $s['addr']);
		$client->hset($HOST, 'i', $UID);
		$client->expire($HOST, 3600 * 3);
	}
}

