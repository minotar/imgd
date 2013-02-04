<?php

class Minotar {

    /**
     * The Default Skin
     */
    const DEFAULT_SKIN = 'char';

    /**
     * Caching On or Off?
     */
    const CACHING = true;

    /**
     * Cache expiration time
     *
     * @var int
     */
    public static $expires = 86400; // 24 hours in seconds

    /**
     * Saves an avatar
     *
     * @param $username
     * @return string $username
     */
    private static function save($username) {
        $mojang = Requests::get('http://s3.amazonaws.com/MinecraftSkins/' . $username . '.png');

        if($mojang->status_code === 200 ) {
            // Good image, let's store
            $img = WideImage::load($mojang->body);

            $img->saveToFile('./minecraft/skins/' . strtolower($username) . '.png');
            $img->destroy();

            return strtolower($username);
        } else {
            // This saves a request
            // Char should already be saved.

            return 'char';
        }
    }

    /**
     * Returns an image skin
     *
     * @param $username
     * @return WideImage_Image
     */
    public static function load($username) {
        if(self::exists($username)) {
            $username = strtolower($username);
            return WideImage::load('./minecraft/skins/'.$username.'.png');
        } else {
            return WideImage::load('./minecraft/skins/'.self::save($username).'.png');
        }
    }

    /**
     * @param $username
     * @return bool
     */
    public static function delete($username) {
        return unlink('./minecraft/skins/'.$username.'.png');
    }


    /**
     * Get and return a skin
     *
     * @param WideImage_Image $resource
     * @return WideImage_Image
     */
    public static function head(WideImage_Image $resource) {
        return $resource->crop(8, 8, 8, 8);
    }

    /**
     * Get and return a helm
     * Returns a head if no helm is found
     *
     * @param WideImage_Image $resource
     * @return WideImage_Image
     */
    public static function helm(WideImage_Image $resource) {
        $background = $resource->getColorAt(0,0);
        $helm = false;

        for ($i = 1; $i <= 8; $i++) {
            for ($j = 1; $j <= 4; $j++) {
                if($resource->getColorAt(40 + $i, 7 + $j) != $background) {
                    $helm = true;
                }
            }

            if($helm) {
                break;
            }
        }

        if($helm) {
            $top = $resource->crop(40, 8, 8, 8);
            $head = self::head($resource);
            $helm = $head->merge($top);
        } else {
            $helm = self::head($resource);
        }

        return $helm;
    }

    /**
     * Checks to see if skin exists
     *
     * @param $username
     * @return bool
     */
    public static function exists($username) {
        if(self::CACHING === false) return false;
        $username = strtolower($username);

        if(file_exists('./minecraft/skins/' . $username . '.png')) {
            return true;
        } else {
            return false;
        }
    }

}
