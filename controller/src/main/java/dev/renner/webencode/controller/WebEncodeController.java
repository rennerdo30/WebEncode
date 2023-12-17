package dev.renner.webencode.controller;

import io.javalin.Javalin;

public class WebEncodeController {
    public static void main(String[] args) {
        var app = Javalin.create(/*config*/)
                .get("/", ctx -> ctx.result("Hello World"))
                .start(8686);
    }
}