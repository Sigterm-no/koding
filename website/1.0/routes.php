<?php

require_once 'router.php';
require_once 'helpers.php';
require_once 'kitecontroller.php';

$router = new Router;

$router->add_route('/kite/:kite_name', function ($params) {
  global $respond;
  $kite_controller = get_kite_controller();
  //header('Access-Control-Allow-Origin: *');
  header('Content-type: text/javascript');
  $session = require_valid_session();
  $uri = $kite_controller->get_kite_uri($params->kite_name, $session['username']);
  if(!isset($uri)) {
    header('HTTP/1.0 404 Not Found');
    $respond(array('error' => 404));
  }
  else {
    $args = $_REQUEST;
    if (isset($session)) {
      $args['username'] = $session['username'];
      $res = @file_get_contents($uri.'?'.http_build_query($args));
      if ($res) {
        $respond($res);
      }
      else {
        header('HTTP/1.0 503 Service Unavailable');
        $respond(array('error' => 503, 'uri' => $uri));
      }      
    }
    else {
      access_denied();
    }
  }
});

$router->add_route('/event', function () {
  @$message = json_decode(file_get_contents('php://input'));
  // error_log(json_encode($message));
  if(isset($message)) {
    foreach ($message->events as $event) {
      switch($event->name) {
      case 'channel_vacated' :
        $matches = array();
        if(preg_match('/^private-(\w+)-/', $event->channel, $matches)) {
          list(, $channel_type) = $matches;
          handle_vacated_channel($channel_type, $event, $message->time_ms);
        }
        break;
      }
    }
  }
  okay();
});

$router->add_route('/login', function () {
  $nonce = $_GET['n'];
  if (!isset($nonce)) {
    access_denied(1);
  }
  $db = get_mongo_db();
  $session = $db->jSessions->findOne(array('nonces' => $nonce));
  
  error_log('session '.var_export($session, TRUE));
  
  if (!isset($session)) {
    access_denied(3);
  }
  $db->jSessions->update(array(
    'nonces' => $nonce,
  ), array(
    '$pull' => array(
      'nonces' => $nonce,
    ),
  ));
  $token = get_token($session);
  if (!isset($token)) {
    access_denied(4);
  }
  setcookie('clientId', $token['token'], $token['expires']->sec);
  okay();
});

$router->add_route('/logout', function () {
  setcookie('clientId', '', time()-3600);
  okay();
});

$router->add_route('/kite/connect', function () {
  $kite_controller = get_kite_controller();
  $req = json_decode($_REQUEST['data']);
  $response = $kite_controller->add_kite($req->kiteName, $req->uri);
  if (!$response) {
    access_denied();
  }
  else okay();
});

$router->add_route('/kite/disconnect', function () {
  # TODO: this is a temporary measure to protect the API until we have real
  # api keys.  Replace with something stronger.
  $secret     = '8daafc24b27ab396d32751f6a8cf2964';
  $token      = $_REQUEST['token'];
  $uri        = $_REQUEST['uri'];
  $kite_name  = $_REQUEST['kiteName'];
  
  if ($token != sha1($uri.$secret)) {
    error_log('unauthorized attempt to send kite/disconnect.');
    access_denied();
  }
  else {
    $kite_controller = get_kite_controller();
    $kite_controller->remove_kite($kite_name, $uri);
    okay();
  }
});
