### Overview
After using [death](http://github.com/vrecan/death) for awhile I started to run into issues closing objects. Mainly that some poorly written code would never close and the app would hang and require me to force kill the app. This seemed like a perfect oppurtunity to use go routines and channels to solve a relatively complex threading problem. What I really want is parallel shutdown with an overall timeout that will shutdown even if something fails. 

[Link to github project](http://github.com/vrecan/death "Application shutdown library for golang").

### How does it work?

So lets look at the new Wait for death method

```go
//Wait for death and then kill all items that need to die.
func (d *Death) WaitForDeath(closable ...Closable) {
	d.wg.Wait()
	log.Info("Shutdown started...")
	count := len(closable)
	log.Debug("Closing ", count, " objects")
	if count > 0 {
		d.closeInMass(closable...)
	}
}
```
What you will see here is that now instead of calling close we are now calling this close in mass method.

```go
//Close all the objects at once and wait forr them to finish with a channel.
func (d *Death) closeInMass(closable ...Closable) {
	count := len(closable)
	//call close async
	done := make(chan bool, count)

	for _, c := range closable {
		go d.closeObjects(c, done)
	}

```
Close in mass is going to get the size of the slice and then make a channel to receive done messages. This channel will then be passed to a goroutine that sends a message when close is finished.

Here is what the function looks like
```go
//Close objects and return a bool when finished on a channel.
func (d *Death) closeObjects(c Closable, done chan<- bool) {
	c.Close()
	done <- true
}
```

Now we use a timer and a select statement to wait for events on the timer channel and the done channel.

```golang
	//wait on channel for notifications.

	timer := time.NewTimer(d.timeout)
	for {
		select {
		case <-timer.C:
			log.Warn(count, " object(s) remaining but timer expired.")
			return
		case <-done:
			count--
			log.Debug(count, " object(s) left")
			if count == 0 {
				log.Debug("Finished closing objects")
				return
			}
		}
	}
}
```
This will block until our timer event is triggered or the number of goroutines running equals 0. 

# Summary

With this change my applications have seen a 2-3x shutdown speed increase with the benefit of always shutting down even if a bad objects close method never returns.

[Managing death in go](http://github.com/vrecan/death "Application shutdown library for golang").


