<?php
// Sets $pkg_root to the root of this package (no trailing slash.)

// Set default value first
$pkg_root = "";

// Check different superglobal keys for different server setups.
if (ISSET($_SERVER['PHP_SELF'])) {
    $pkg_root = dirname($_SERVER['PHP_SELF']);
}
// Add more configurations as necessary.