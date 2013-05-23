<?php
require 'pkg_root.php';
?>
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <title><?php echo (isset($this->title)) ? $this->title : 'The Miner\'s Avatar'; ?> &mdash; Minotar</title>

        <!-- Meta -->
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta name="description" content="A Minotar is a global avatar that pulls your head off of your Minecraft.net skin, and allows it for use on thousands of sites. See some uses below.">
        <meta name="author" content="Axxim, LLC">

        <!-- Style -->
        <!--<link href="./assets/css/style.css" rel="stylesheet">-->
        <link rel="stylesheet" href="<?php echo $pkg_root; ?>/assets/css/base.css">
        <link rel="stylesheet" href="<?php echo $pkg_root; ?>/assets/css/skeleton.css">
        <link rel="stylesheet" href="<?php echo $pkg_root; ?>/assets/css/layout.css">
        <link rel="stylesheet" href="<?php echo $pkg_root; ?>/assets/css/minotar.css">

        <!--[if lt IE 9]>
            <script src="//html5shim.googlecode.com/svn/trunk/html5.js"></script>
        <![endif]-->

        <link rel="shortcut icon" href="<?php echo $pkg_root; ?>/avatar/clone1018/128.png">

        <script>
            var _gaq = _gaq || [];
            _gaq.push(['_setAccount', 'UA-22574465-1']);
            _gaq.push(['_trackPageview']);

            (function() {
                var ga = document.createElement('script'); ga.async = true;
                ga.src = ('https:' == document.location.protocol ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js';
                var s = document.getElementsByTagName('script')[0]; s.parentNode.insertBefore(ga, s);
            })();
        </script>
    </head>
    <body>
        <a href="https://github.com/axxim/minotar"><img class="fork" src="https://s3.amazonaws.com/github/ribbons/forkme_left_darkblue_121621.png" alt="Fork me on GitHub"></a>
        <div class="container">
            <div class="sixteen columns">
                <?php $this->yield(); ?>
            </div>
        </div>
        <script type="text/javascript">var RumID = 40;</script><script type="text/javascript" src="https://statuscake.com/App/Workfloor/RUM/Embed.js"></script>
    </body>
</html>