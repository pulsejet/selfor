<?php
require 'common.php';

session_start();

$UID = "someone";

// Get params
$HOST = $_SERVER['HTTP_X_FORWARDED_FOR'];
$PSK = $_GET["psk"];
$CN = $_GET["cn"];

// Parse server object
function parse($r) {
	$s = array();
	$s["psk"] = $r[0];
	$s["addr"] = $r[1];
	$s["name"] = $r[2];
	$s["cn"] = $r[3];
	return $s;
}

$PUSH = null;

$servers = array();

$file = fopen("servers.csv","r");
while(!feof($file)) {
	$s = fgetcsv($file);
	if (sizeof($s) != 4) continue;
	$s = parse($s);

	if ($PSK != "" && $s["psk"] == $PSK) {
		array_push($servers, $s);
		if ($CN != "" && $s["cn"] == $CN) {
			$PUSH = $s;
			push($s, $HOST, $UID);
			$_SESSION['ip'] = $HOST;
		}
	}
}
fclose($file);

?>
<!DOCTYPE HTML>
<html>
<head>
	<title> SSH Access </title>
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style>
	* { font-family: monospace; }
	input[type=text] {
		border: 2px solid gray;
		border-radius: 4px;
		padding: 7px 12px;
	}
	input[type=submit], input[type=reset] {
		background-color: #4CAF50;
		border: none;
		border-radius: 4px;
		color: white;
		padding: 8px 12px;
		text-decoration: none;
		margin: 4px 2px;
		cursor: pointer;
	}
	</style>
</head>

<body>
	<h1> SSH Access </h1>

	<h2> Authorized as <?php echo $UID; ?> </h2>
	<form method="GET" action="logout.php">
		<input type=submit value="Logout" />
	</form>

	<?php if ($PUSH != NULL) {?>

	<h2> Lease obtained for <?php echo $PUSH["name"] . " (" . $PUSH["addr"] . ")"; ?> </h2>
	<h2> <span style="color: red"> ssh -o "CheckHostIP no" -p 443 <?php echo $PUSH['cn']; ?>.ssh.domain.com </span> </h2>
	<p>
		Your lease will last for 3 hours. Refresh this page to renew the lease. <br/>
		Active connections will not be disconncted if your lease expires
	</p>

	<?php } ?>

	<h2> Enter Pre-Shared Key </h2>
	<form method="GET">
		<input name="psk" placeholder="PSK" value="<?php echo $PSK ?>" type="text" />
		<input type=submit value="Submit" />
	</form>

	<h2> Select Server [Your IP: <?php echo $HOST ?>] </h2>
	<table>
	<?php
	foreach ($servers as $s) { ?>
	<tr>
		<td style="font-size: 1.2em">
			<?php echo $s["name"] . " (" . $s["addr"] . ")"; ?>
		</td>
		<td>
			<form method="GET">
				<input name="psk" placeholder="PSK" value="<?php echo $PSK ?>" type="text" hidden />
				<input name="cn" value="<?php echo $s["cn"] ?>" type="text" hidden />
				<input type=submit value="Select" />
			</form>
		</td>
	</tr>
	<?php } ?>
	</table>
</body>
</html>

