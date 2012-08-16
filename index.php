<?php
require 'klein.php';
include 'WideImage/WideImage.php';
include 'Minotar.php';

error_reporting(0);

respond('/', function ($request, $response) {
    $response->render('html/home.phtml');
});

respond('/[avatar|head]/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext = $request->param('format', '.png');
    $ext = $request->param('formate', '.png');
    list($name) = explode('.', $name);
    list($size) = explode('.', $size);
    $size = min(1000, max(1, (int) $size));

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
    $size = min(1000, max(1, (int) $size));

    $name = Minotar::get($name);
    
    $head = WideImage::load("./minecraft/heads/$name.png")->resize($size);
    $helm = WideImage::load("./minecraft/helms/$name.png")->resize($size);
    $pixel = $helm->getColorAt(1,1);
    $black = array('red' => 0, 'green' => 0, 'blue' => 0, 'alpha' => 0);
    if($helm->getColorRGB($pixel) == $black)
        $result = clone $head;
    else
        $result = $head->merge($helm);
    
    $result->output($ext);
});

respond('/[player|body]/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext = $request->param('format', '.png');
    $ext = $request->param('formate', '.png');
    list($name) = explode('.', $name);
    list($size) = explode('.', $size);
    $size = min(1000, max(1, (int) $size));

    $name = Minotar::get($name);

    $img = WideImage::load("./minecraft/players/$name.png")->resize($size);
    $img->output($ext);
});

respond('/random/[:size]?.[:format]?', function ($request, $response) {
    $size = $request->param('size', 180);
	$ext  = $request->param('format', '.png');
    list($size) = explode('.', $size);
    $size = min(1000, max(1, (int) $size));

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

respond('/all/[:type]?', function ($request, $response) {
    $type = $request->param('type', 'heads');
    $files = Minotar::getFilesFromDir("./minecraft/$type");
    $response->render('html/all.phtml', array('files' => $files));
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

respond('404', function ($request, $response) {
    $response->render('html/404.phtml');
});

dispatch();
