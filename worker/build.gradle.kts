plugins {
    id("java")
}

group = "dev.renner.webencode.worker"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {

    // https://mvnrepository.com/artifact/net.bramp.ffmpeg/ffmpeg
    implementation("net.bramp.ffmpeg:ffmpeg:0.8.0")


    testImplementation(platform("org.junit:junit-bom:5.9.1"))
    testImplementation("org.junit.jupiter:junit-jupiter")
}

tasks.test {
    useJUnitPlatform()
}