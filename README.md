# Minotar

A Minotar is a global avatar that pulls your head off your Minecraft.net skin, and allows it for use on several thousand sites - anywhere you can embed an image. See some uses below.

![clone1018](https://minotar.net/avatar/clone1018/64)
![citricsquid](https://minotar.net/avatar/citricsquid/64)
![Raitsui](https://minotar.net/avatar/Raitsui/64)
![runforthefinish](https://minotar.net/avatar/runforthefinish/64)
![NoMercyJon](https://minotar.net/avatar/NoMercyJon/64)
![Nautika](https://minotar.net/avatar/Nautika/64)
![Notch](https://minotar.net/avatar/Notch/64)
![NiteAngel](https://minotar.net/helm/NiteAngel/64)
![hyp3rdriv3](https://minotar.net/helm/hyp3rdriv3/64)
![S1NZ](https://minotar.net/helm/S1NZ/64)
![KakashiSuno](https://minotar.net/helm/KakashiSuno/64)

## Sweet and Simple API

### Simple Heads
Unlike the PayPal API, we keep things nice and simple. For basic usage just provide a username:
`<img src="https://minotar.net/avatar/clone1018">`

You can also set a size. We use pixels and we only need the width. Just add it to the end.
`<img src="https://minotar.net/avatar/clone1018/100">`

And since some services require an extension we've added simple support for it. Just add .png to the end.
`<img src="https://minotar.net/avatar/clone1018/100.png">`

### Avatar With Helm
Sometimes you want to display a helm too, that's fine with this endpoint.
`<img src="https://minotar.net/helm/clone1018/100.png">`


## Advanced

#### User's Skin
You can even use Minotar's API to get a users skin. We're adding more soon!
`<img src="https://minotar.net/skin/clone1018">`

You can also set the browser to download the image by using:
`https://minotar.net/download/clone1018`

#### Default Skin
Need Steve? Use "char" as the username:

`<img src="https://minotar.net/skin/char">`

## How to install?
- Coming Soon -

## Understanding Headers
We use a couple of headers to help in understanding how something is served, here they are:

x-requested

- returns: processed
- explain: if Minotar processed your avatar

x-result:

- returns: ok,failed
- explain: ok on successful GET from s3, failed on failed GET from s3

x-timing:

- returns: fetch time, process time, resize time, whole process
- example: 48+0+4=52

## Thanks
Big thanks to [lukegb](https://github.com/lukegb) for porting the old version of this script from PHP to Go.
