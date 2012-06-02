<?php require 'klein.php';
include 'WideImage/WideImage.php';
include 'Minotar.php';
error_reporting(0);

respond('/', function ($request, $response) {
    $response->render('html/home.phtml');
});

respond('/[avatar|head]/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext  = $request->param('format', '.png');
    $ext  = $request->param('formate', '.png');
    $name = explode('.', $name); $name = $name[0];
    $size = explode('.', $size); $size = $size[0];
    if($size >= 1000) $size = 1000;

    $name = Minotar::get($name);

    $img = WideImage::load("./minecraft/skins/$name.png")->crop(8,8,8,8)->resize($size);
    $img->output($ext);
});

respond('/helm/[:username].[:format]?/[:size]?.[:formate]?', function ($request, $response) {
    $name = $request->param('username', 'char');
    $size = $request->param('size', 180);
    $ext  = $request->param('format', '.png');
    $ext  = $request->param('formate', '.png');
    $name = explode('.', $name); $name = $name[0];
    $size = explode('.', $size); $size = $size[0];
    if($size >= 1000) $size = 1000;

    $name = Minotar::get($name);
	$black = 'iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAYAAADED76LAAAAFklEQVQY02NkYGD4z4AHMDEQAMNDAQAMUwEPqfUHaQAAAABJRU5ErkJggg=='; // Black 16x16 for comparison
    $watermark = WideImage::load("./minecraft/skins/$name.png")->crop(40,8,8,8);
	$base = WideImage::load("./minecraft/skins/$name.png")->crop(8,8,8,8);
	if (base64_encode($watermark) != $black) {
		$result = $base->merge($watermark);
	} else {
		$result = $base;
	}

    $result->resize($size)->output($ext);
});

respond('/random/[:size]?.[:format]?', function ($request, $response) {
	$size = $request->param('size', 180);
	$ext  = $request->param('format', '.png');
	$size = explode('.', $size); $size = $size[0];
	if($size >= 1000) $size = 1000;

	$avatars = scandir('./minecraft/heads/');
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
	header('Content-Disposition: attachment; filename="'.$name.'.png"');
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
	if($width >= 1920) $width = 1920;
	if($height >= 1080) $height = 1080;
	$files = array_slice($files, 500);

	//list($width, $height) = getimagesize($_GET['image']);
    $image_p = imagecreatetruecolor($width, $height);
    $count = 1;
    foreach($files as $avatar) {
    	$image = imagecreatefrompng($avatar);
    	imagecopyresampled($image_p, $image, $width * $count, $height * $count, 0, 0, $width, $height, 42, 42);
    	$count++;
    }
    header('Content-type: image/png');
    imagejpeg($image_p, null, 100);
});

dispatch(); 