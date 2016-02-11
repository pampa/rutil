var gulp = require('gulp');  
var exec = require("child_process").exec;

gulp.task('install', function() {
  exec("go install",   function (error, stdout, stderr) {
    if (stdout !=  "") { console.log(stdout); }
    if (stderr !=  "") { console.log(stderr); }
    if (error  !== null) { console.log('exec error: ' + error); }
  });
});

gulp.task('watch', function() {  
    gulp.watch('**/*.go', ['install']);
});

gulp.task('default', ['watch']);
