<?php
require 'vendor/autoload.php';
require 'vendor/portable.php';

require 'minotar/Minotar.php';

$pkg_root = dirname($_SERVER['PHP_SELF']);

respond($pkg_root . '/', function ($request, $response) {

    $response->layout('html/template/default.php');
    $response->render('html/home.php');

});

respond($pkg_root . '/tests', function ($request, $response) {

    $response->layout('html/template/default.php');
    $response->render('html/tests.php');

});

addValidator('username', function ($str) {
    return preg_match('/^[a-z0-9_-]+$/i', $str);
});

respond($pkg_root . '/[avatar|head]/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $request->validate('username')->isUsername();

    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext = $request->param('format', 'png');
    $ext = $request->param('formate', 'png');
    list($name) = explode('.', $name);
    list($size) = explode('.', $size);
    $size = min(1000, max(16, (int)$size));

    $skin = Minotar::load($name);
    $head = Minotar::head($skin);

    $img = $head->resize($size);
    $img->output($ext);
});

respond($pkg_root . '/helm/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $request->validate('username')->isUsername();

    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext = $request->param('format', 'png');
    $ext = $request->param('formate', 'png');
    list($name) = explode('.', $name);
    list($size) = explode('.', $size);
    $size = min(1000, max(16, (int)$size));

    $skin = Minotar::load($name);
    $helm = Minotar::helm($skin);

    $helm->resize($size)->output($ext);
});

respond($pkg_root . '/[player|body]/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $request->validate('username')->isUsername();

    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext = $request->param('format', 'png');
    $ext = $request->param('formate', 'png');
    list($name) = explode('.', $name);
    list($size) = explode('.', $size);
    $size = min(1000, max(16, (int)$size));

    $skin = Minotar::load($name);

    $skin->output($ext);
});

respond($pkg_root . '/random/[:size]?.[:format]?', function ($request, $response) {
    $size = $request->param('size', 180);
    $ext = $request->param('format', 'png');
    list($size) = explode('.', $size);
    $size = min(1000, max(16, (int)$size));

    // We're phasing this out, load me :D
    $skin = Minotar::load('clone1018');
    $head = Minotar::head($skin);

    $img = $head->resize($size);
    $img->output($ext);
});

respond($pkg_root . '/download/[:username]', function ($request, $response) {
    $request->validate('username')->isUsername();

    $name = $request->param('username', 'char');

    $skin = Minotar::get($name);
    header('Content-Disposition: attachment; filename="' . $name . '.png"');
    $skin->output('.png');
});

respond($pkg_root . '/skin/[:username]', function ($request, $response) {
    $request->validate('username')->isUsername();

    $name = $request->param('username', 'char');

    $skin = Minotar::load($name);
    $skin->output('png');
});

respond($pkg_root . '/wallpaper/[:width]/[:height]?', function ($request, $response) {
    $width = $request->param('width', 1024);
    $height = $request->param('height', 768);

    // In development
});

respond($pkg_root . '/refresh/[:username]', function ($request, $response) {
    $request->validate('username')->isUsername();

    $username = $request->param('username');
    $name = Minotar::delete($username);
    //Header("Location: /avatar/$username");
});

respond('404', function ($request, $response) {
    $response->layout('html/template/default.php');
    $response->render('html/404.php', array("title" => "404"));
});

dispatch();
