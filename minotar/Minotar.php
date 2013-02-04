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
     * Get and save a Minotar
     *
     * @param string $username
     * @param bool $clear
     */
    public static function save($username, $clear = false) {
        $mojang = Requests::get('http://s3.amazonaws.com/MinecraftSkins/' . $username . '.png');

        if($mojang->status_code === 200) {
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

    public static function load($username) {
        $username = strtolower($username);

        return WideImage::load('./minecraft/skins/'.$username.'.png');
    }


    public static function head($resource) {
        return $resource->crop(8, 8, 8, 8);
    }



    /**
     * See if an avatar exists
     *
     * @param string $username
     */
    public static function exists($username) {
        if(self::CACHING === FALSE) return false;
        $username = strtolower($username);

        if(file_exists('./minecraft/skins/' . $username . '.png')) {
            return true;
        } else {
            return false;
        }
    }

}
