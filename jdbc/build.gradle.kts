plugins {
    `java-library`
    `maven-publish`
}

group = "io.github.mystisql"
version = "1.1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
    withSourcesJar()
}

sourceSets {
    main {
        java {
            exclude("**/examples/**")
        }
    }
}

dependencies {
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("org.java-websocket:Java-WebSocket:1.5.4")
    implementation("com.fasterxml.jackson.core:jackson-databind:2.16.1")
    implementation("com.fasterxml.jackson.datatype:jackson-datatype-jsr310:2.16.1")
    implementation("org.slf4j:slf4j-api:1.7.36")
    
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.1")
    testImplementation("org.mockito:mockito-core:5.8.0")
    testImplementation("com.squareup.okhttp3:mockwebserver:4.12.0")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

tasks.test {
    useJUnitPlatform()
}

tasks.jar {
    manifest {
        attributes(
            "Implementation-Title" to "MystiSql JDBC Driver",
            "Implementation-Version" to project.version,
            "Implementation-Vendor" to "MystiSql",
            "Automatic-Module-Name" to "io.github.mystisql.jdbc"
        )
    }
}

val fatJar = tasks.register("fatJar", Jar::class) {
    group = "build"
    archiveClassifier.set("all")
    manifest {
        attributes(
            "Implementation-Title" to "MystiSql JDBC Driver (Shaded)",
            "Implementation-Version" to project.version,
            "Implementation-Vendor" to "MystiSql",
            "Automatic-Module-Name" to "io.github.mystisql.jdbc"
        )
    }
    duplicatesStrategy = DuplicatesStrategy.EXCLUDE
    
    dependsOn(configurations.runtimeClasspath)
    from(sourceSets.main.get().output)
    configurations.runtimeClasspath.get().forEach { dep ->
        if (dep.isDirectory) {
            from(dep)
        } else {
            from(zipTree(dep))
        }
    }
    exclude("META-INF/*.SF", "META-INF/*.DSA", "META-INF/*.RSA")
}

tasks.build {
    dependsOn(fatJar)
}

publishing {
    publications {
        create<MavenPublication>("mavenJava") {
            from(components["java"])
            pom {
                name.set("MystiSql JDBC Driver")
                description.set("JDBC driver for MystiSql - Access K8s databases transparently")
                url.set("https://github.com/mystisql/mystisql")
                licenses {
                    license {
                        name.set("The Apache License, Version 2.0")
                        url.set("http://www.apache.org/licenses/LICENSE-2.0.txt")
                    }
                }
                developers {
                    developer {
                        id.set("mystisql")
                        name.set("MystiSql Team")
                        email.set("mystisql@example.com")
                    }
                }
                scm {
                    connection.set("scm:git:git://github.com/mystisql/mystisql.git")
                    developerConnection.set("scm:git:ssh://github.com/mystisql/mystisql.git")
                    url.set("https://github.com/mystisql/mystisql")
                }
            }
        }
    }
}
