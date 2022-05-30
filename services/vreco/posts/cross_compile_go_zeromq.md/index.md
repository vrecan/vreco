I've found it really dificult to find a concrete tutorial for building zeromq for go in windows using MinGW in centos 7. First lets add the epel-release repo.</p>

```bash
sudo yum install epel-release
```

<p>Now lets install mingw64. For this tutorial I am going to just install all the mingw libraries.</p>

```bash
sudo yum install mingw64*
```

<p> Now lets download and compile zeromq using mingw.</p>

```bash
wget http://download.zeromq.org/zeromq-4.0.5.zip
unzip zeromq-4.0.5.zip ~/zeromq
cd ~/zeromq
mingw64-configure configure
mingw64-make
```

<p>Now lets recompile go so that it is also using the mingw compiler.</p>

```bash
cd /usr/local/go/src
sudo  env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC_FOR_TARGET="x86_64-w64-mingw32-gcc" ./make.bash
```

<p>Lets move to the go project that you are going to be using with zeromq.</p>

```bash
cd <PROJECT DIR>
env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC_FOR_TARGET="x86_64-w64-mingw32-gcc" CGO_LDFLAGS="-L/home/<username>/zeromq/src/.libs -static-libgcc" go build -v -x  -o app.exe
minigw_bin=/usr/x86_64-w64-mingw32/sys-root/mingw/bin/
```

<p>This should have created an exe binary that is linked to the mingw dll's and your recently compiled version of zeromq. Now lets copy the dll's for the required libraries to the current directory so that we can package this up and put it on a windows box.</p>

```bash
cp $minigw_bin/libwinpthread-1.dll .
cp $minigw_bin/libgcc_s_seh-1.dll .
cp $minigw_bin/libstdc++-6.dll .
cp ~/zeromq/src/.libs/libzmq.dll .
```

<p> If you now transfer this folder to a windows(64) 7/8/8.1 box you can now just execute the .exe and have a working go windows application using zeromq.
