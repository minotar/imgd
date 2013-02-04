<?php
require 'vendor/autoload.php';
require 'minotar/Minotar.php';

respond('/', function ($request, $response) {

    $response->layout('html/template/default.php');
    $response->render('html/home.php');

});

respond('/tests', function ($request, $response) {

    $response->layout('html/template/default.php');
    $response->render('html/tests.php');

});

addValidator('username', function ($str) {
    return preg_match('/^[a-z0-9_-]+$/i', $str);
});

respond('/[avatar|head]/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
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

respond('/helm/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
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

respond('/[player|body]/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
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

respond('/random/[:size]?.[:format]?', function ($request, $response) {
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

respond('/download/[:username]', function ($request, $response) {
    $request->validate('username')->isUsername();
    
    $name = $request->param('username', 'char');

    $skin = Minotar::get($name);
    header('Content-Disposition: attachment; filename="' . $name . '.png"');
    $skin->output('.png');
});

respond('/skin/[:username]', function ($request, $response) {
    $request->validate('username')->isUsername();
    
    $name = $request->param('username', 'char');

    $skin = Minotar::load($name);
    $skin->output('png');
});

respond('/wallpaper/[:width]/[:height]?', function ($request, $response) {
    $width = $request->param('width', 1024);
    $height = $request->param('height', 768);

    // In development
});

respond('/refresh/[:username]', function ($request, $response) {
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
