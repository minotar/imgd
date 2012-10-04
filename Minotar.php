<?php

/**
 *
 */
class Minotar {

    /**
     * The Default Skin
     */
    const DEFAULT_SKIN = 'char';
    /**
     * Caching On or Off?
     */
    const CACHING = false;

    /**
     * Cache expiration time
     *
     * @var int
     */
    public static $expires = 86400; // 24 hours in seconds

    /**
     * Downloads a minecraft skin from Mojang
     *
     * @param string $username
     * @param bool $clear
     * @return string
     */
    public static function get($username, $clear = false) {
        if (!file_exists('./minecraft/skins/' . strtolower($username) . '.png') || $clear) {
            $contents = self::fetch('http://s3.amazonaws.com/MinecraftSkins/' . $username . '.png');
            if ($contents === false) {
                $img = WideImage::load("http://s3.amazonaws.com/MinecraftSkins/char.png");
                $img->saveToFile("./minecraft/skins/char.png");
                $helm = clone $img;
                $helm->crop(40, 8, 8, 8)->saveToFile('./minecraft/helms/char.png');
                $head = clone $img;
                $head->crop(8, 8, 8, 8)->saveToFile('./minecraft/heads/char.png');

                $head->destroy();
                $helm->destroy();
                $img->destroy();

                header("Status: 404 Not Found");
                return 'char';
            } else {
                try {
                    $img = WideImage::load($contents);
                } catch (Exception $e) {
                    return 'char';
                }

                $img->saveToFile('./minecraft/skins/' . strtolower($username) . '.png');
                $helm = clone $img;
                $helm->crop(40, 8, 8, 8)->saveToFile('./minecraft/helms/' . strtolower($username) . '.png');
                $head = clone $img;
                $head->crop(8, 8, 8, 8)->saveToFile('./minecraft/heads/' . strtolower($username) . '.png');

                $head->destroy();
                $helm->destroy();
                $img->destroy();

                return strtolower($username);
            }
        } else {
            return strtolower($username);
        }
    }

    /**
     * Gets the files from a directory
     *
     * @param $dir
     * @return array
     */
    public static function getFilesFromDir($dir) {
        $files = array();
        if ($handle = opendir($dir)) {
            while (false !== ($file = readdir($handle))) {
                if ($file != "." && $file != ".." && $file != "Thumbs.db") {
                    if (is_dir($dir . '/' . $file)) {
                        $dir2 = $dir . '/' . $file;
                        $files[] = getFilesFromDir($dir2);
                    } else {
                        $files[] = $dir . '/' . $file;
                    }
                }
            }
            closedir($handle);
        }
        return Minotar::array_flat($files);
    }

    /**
     * Flattens an array
     *
     * @param array $array
     * @return array
     */
    private static function array_flat($array) {
        foreach ($array as $a) {
            if (is_array($a)) {
                $tmp = array();
                $tmp = array_merge($tmp, array_flat($a));
            } else {
                $tmp[] = $a;
            }
        }

        return $tmp;
    }

    /**
     * Checks URL to make sure it exists
     *
     * @param string $url
     * @return bool|mixed
     */
    private function found($url) {
        $handle = curl_init($url);
        if (false === $handle) {
            return false;
        }
        curl_setopt($handle, CURLOPT_HEADER, false);
        curl_setopt($handle, CURLOPT_FAILONERROR, true);  // this works
        curl_setopt($handle, CURLOPT_HTTPHEADER, Array("User-Agent: Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.15) Gecko/20080623 Firefox/2.0.0.15")); // request as if Firefox
        curl_setopt($handle, CURLOPT_NOBODY, true);
        curl_setopt($handle, CURLOPT_RETURNTRANSFER, false);
        $connectable = curl_exec($handle);
        curl_close($handle);
        return $connectable;
    }

    /**
     * wget for PHP
     *
     * @param string $url
     * @return bool|mixed
     */
    private static function fetch($url) {
        $handle = curl_init($url);
        if (false === $handle) {
            return false;
        }
        curl_setopt($handle, CURLOPT_HEADER, false);
        curl_setopt($handle, CURLOPT_FAILONERROR, true);  // this works
        curl_setopt($handle, CURLOPT_HTTPHEADER, Array("User-Agent: Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.15) Gecko/20080623 Firefox/2.0.0.15")); // request as if Firefox
        curl_setopt($handle, CURLOPT_RETURNTRANSFER, true);
        $contents = curl_exec($handle);
        curl_close($handle);
        return $contents;
    }

}
