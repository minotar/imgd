<?php

class Minotar {

    const DEFAULT_SKIN = 'char';
    const CACHING = true;

    public static $expires = 86400; // 24 hours in seconds

    public static function get($username) {
        if (!file_exists('./minecraft/skins/' . strtolower($username) . '.png')) {
            $contents = self::fetch('http://s3.amazonaws.com/MinecraftSkins/' . $username . '.png');
            if ($contents === false) {
                $img = WideImage::load("http://s3.amazonaws.com/MinecraftSkins/char.png");
                $img->saveToFile("./minecraft/skins/char.png");
                $helm = clone $img;
                $helm->crop(40, 8, 8, 8)->saveToFile('./minecraft/helms/char.png');
                $head = clone $img;
                $head->crop(8, 8, 8, 8)->saveToFile('./minecraft/heads/char.png');
                header("Status: 404 Not Found");
                return 'char';
            } else {
                $img = WideImage::load($contents);
                $img->saveToFile('./minecraft/skins/' . strtolower($username) . '.png');
                $helm = clone $img;
                $helm->crop(40, 8, 8, 8)->saveToFile('./minecraft/helms/' . strtolower($username) . '.png');
                $head = clone $img;
                $head->crop(8, 8, 8, 8)->saveToFile('./minecraft/heads/' . strtolower($username) . '.png');
                return strtolower($username);
            }
        } else {
            return strtolower($username);
        }
    }

    public function getFilesFromDir($dir) {
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

    private function array_flat($array) {
        foreach ($array as $a) {
            if (is_array($a)) {
                $tmp = array_merge($tmp, array_flat($a));
            } else {
                $tmp[] = $a;
            }
        }

        return $tmp;
    }

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
