<div class="twelve columns content">
    <h1><?php echo $this->type; ?></h1>
    <p class="center">
        View all: <a href="<?php echo URL; ?>all/head">heads</a> / <a href="<?php echo URL; ?>all/helm">helms</a> / <a href="<?php echo URL; ?>all/skin">skins</a>
    </p>
    <p class="center">
        <?php if($this->start > 0) { ?><a href="<?php echo URL; ?>all/<?php echo $this->type; ?>/<?php echo ($this->start - $this->limit); ?>">&larr; Previous <?php echo $this->limit; ?></a><?php } ?>
        Viewing <?php echo $this->start; ?> - <?php echo ($this->start + $this->limit); ?> of <?php echo $this->total; ?>
        <?php if($this->start + $this->limit < $this->total ) { ?>&mdash; <a href="<?php echo URL; ?>all/<?php echo $this->type; ?>/<?php echo ($this->start + $this->limit); ?>">view next <?php echo $this->limit; ?> &rarr;</a><?php } ?>
    </p>
    <ul class="heads">
        <?php $urls = array("skin" => "skin/%username%", "head" => "head/%username%/64", "helm" => "helm/%username%/64"); ?>
        <?php foreach($this->files as $username) { ?>
        <li class="<?php echo $this->type; ?>"><img src="<?php echo URL; ?><?php echo str_replace("%username%", $username, $urls[$this->type]); ?>" alt="<?php echo $username; ?>" title="<?php echo $username; ?>"/></li>
        <?php } ?>
    </ul>
    <div style="clear:both"></div>
    <p class="center"><a href="<?php echo URL; ?>">&larr; Return to Minotar</a></p>
</div>
