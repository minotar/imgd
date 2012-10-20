<?php
require 'klein.php';
include 'WideImage/WideImage.php';
include 'Minotar.php';

define('URL', 'https://minotar.net/');
//error_reporting(0);

respond('/', function ($request, $response) {

    $response->layout('html/template/default.php');
    $response->render('html/home.php');

});

respond('/[avatar|head]/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext = $request->param('format', '.png');
    $ext = $request->param('formate', '.png');
    list($name) = explode('.', $name);
    list($size) = explode('.', $size);
    $size = min(1000, max(16, (int) $size));

    $name = Minotar::get($name);

    $img = WideImage::load("./minecraft/heads/$name.png")->resize($size);
    $img->output($ext);
});

respond('/helm/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext = $request->param('format', '.png');
    $ext = $request->param('formate', '.png');
    list($name) = explode('.', $name);
    list($size) = explode('.', $size);
    $size = min(1000, max(16, (int) $size));

    $name = Minotar::get($name);

    $head = WideImage::load("./minecraft/heads/$name.png")->resize($size);
    $helm = WideImage::load("./minecraft/helms/$name.png")->resize($size);

    if($helm->isTransparent())
        $result = $head->merge($helm);
    else
        $result = clone $head;

    $result->output($ext);
});

respond('/[player|body]/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext = $request->param('format', '.png');
    $ext = $request->param('formate', '.png');
    list($name) = explode('.', $name);
    list($size) = explode('.', $size);
    $size = min(1000, max(16, (int) $size));

    $name = Minotar::get($name);

    $img = WideImage::load("./minecraft/players/$name.png")->resize($size);
    $img->output($ext);
});

respond('/random/[:size]?.[:format]?', function ($request, $response) {
    $size = $request->param('size', 180);
	$ext  = $request->param('format', '.png');
    list($size) = explode('.', $size);
    $size = min(1000, max(16, (int) $size));

    $avatars = array_diff(scandir('./minecraft/heads/'), array('..', '.', '.gitignore'));
    $rand = array_rand($avatars);

    $avatar = $avatars[$rand];

    $name = str_replace('.png', '', $avatar);
    header("Cache-Control: no-cache, must-revalidate");
    header("Expires: Sat, 26 Jul 1997 05:00:00 GMT");
    WideImage::load("./minecraft/heads/$name.png")->resize($size)->output('png');
});

respond('/download/[:username]', function ($request, $response) {
    $name = $request->param('username', 'char');

    $name = Minotar::get($name);

    $img = WideImage::load("./minecraft/skins/$name.png");
    header('Content-Disposition: attachment; filename="' . $name . '.png"');
    $img->output('.png');
});

respond('/skin/[:username]', function ($request, $response) {
    $name = $request->param('username', 'char');

    $name = Minotar::get($name);

    $img = WideImage::load("./minecraft/skins/$name.png");
    $img->output('.png');
});

respond('/all/[head|helm|skin:type]/[i:start]?', function ($request, $response) {
    $type = $request->param('type', 'head');
    $start = $request->param('start', 0);
    $limit = 85;
    $files = Minotar::getFilesFromDir("./minecraft/{$type}s");

    $response->layout('html/template/default.php');

    foreach(array_slice($files, $start, $limit) as $file) {
        $segments = explode("/", $file);
        $dir_list = array_values((explode(".", end($segments))));
        $file_list[] = array_shift($dir_list);
    }

    if($files) {
        $response->render('html/all.php', array('files' => $file_list, 'type' => $type, 'start' => $start, 'limit' => $limit, 'total' => count($files), 'title' => "All {$type}s"));
        return;
    }

    $response->render('html/404.php', array("title" => "404"));
});

respond('/wallpaper/[:width]/[:height]?', function ($request, $response) {
    $width = $request->param('width', 1024);
    $height = $request->param('height', 768);
    $files = Minotar::getFilesFromDir("./minecraft/heads");
    if ($width >= 1920)
        $width = 1920;
    if ($height >= 1080)
        $height = 1080;
    $files = array_slice($files, 500);

    //list($width, $height) = getimagesize($_GET['image']);
    $image_p = imagecreatetruecolor($width, $height);
    $count = 1;
    foreach ($files as $avatar) {
        $image = imagecreatefrompng($avatar);
        imagecopyresampled($image_p, $image, $width * $count, $height * $count, 0, 0, $width, $height, 42, 42);
        $count++;
    }
    header('Content-type: image/png');
    imagejpeg($image_p, null, 100);
});

respond('/refresh/[:username]', function ($request, $response) {
    $username = $request->param('username');
    $name = Minotar::get($username, true);
    Header("Location: ".URL."avatar/$username");
});

respond('404', function ($request, $response) {
    $response->layout('html/template/default.php');
    $response->render('html/404.php', array("title" => "404"));
});

dispatch();
