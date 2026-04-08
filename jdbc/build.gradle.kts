plugins {
    `java-library`
    `maven-publish`
    id("io.github.goooler.shadow") version "8.1.8"
}

group = "io.github.mystisql"
version = project.findProperty("version") as String? ?: "1.1.0"

repositories {
    mavenCentral()
}

java {
    sourceCompatibility = JavaVersion.VERSION_21
    targetCompatibility = JavaVersion.VERSION_21
    withSourcesJar()
    toolchain {
        languageVersion = JavaLanguageVersion.of(21)
    }
}

sourceSets {
    main {
        java {
            exclude("**/examples/**")
        }
    }
    test {
        java {
            srcDirs("../e2e-test/jdbc")
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
    useJUnitPlatform {
        excludeTags("integration")
    }
    maxParallelForks = 1
    systemProperty("junit.jupiter.execution.timeout.default", "30s")
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

tasks.shadowJar {
    archiveClassifier.set("all")
    manifest {
        attributes(
            "Implementation-Title" to "MystiSql JDBC Driver (Shaded)",
            "Implementation-Version" to project.version,
            "Implementation-Vendor" to "MystiSql",
            "Automatic-Module-Name" to "io.github.mystisql.jdbc"
        )
    }
    
    relocate("okhttp3", "io.github.mystisql.shaded.okhttp3")
    relocate("okio", "io.github.mystisql.shaded.okio")
    relocate("com.fasterxml.jackson", "io.github.mystisql.shaded.jackson")
    relocate("org.java_websocket", "io.github.mystisql.shaded.websocket")
    
    exclude("META-INF/*.SF", "META-INF/*.DSA", "META-INF/*.RSA")
    
    minimize {
        exclude(dependency("org.java-websocket:Java-WebSocket:.*"))
        exclude(dependency("com.fasterxml.jackson.core:.*:.*"))
    }
}

tasks.build {
    dependsOn(tasks.shadowJar)
}

publishing {
    publications {
        create<MavenPublication>("mavenJava") {
            from(components["java"])
            artifact(tasks.shadowJar)
            pom {
                name.set("MystiSql JDBC Driver")
                description.set("JDBC driver for MystiSql - Access K8s databases transparently")
                url.set("https://github.com/L-ios/MystiSql")
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
                    }
                }
                scm {
                    connection.set("scm:git:git://github.com/L-ios/MystiSql.git")
                    developerConnection.set("scm:git:ssh://github.com/L-ios/MystiSql.git")
                    url.set("https://github.com/L-ios/MystiSql")
                }
            }
        }
    }
    repositories {
        maven {
            name = "GitHubPackages"
            url = uri("https://maven.pkg.github.com/${project.property("github.repository")}")
            credentials {
                username = project.findProperty("gpr.user") as String? ?: System.getenv("GITHUB_ACTOR")
                password = project.findProperty("gpr.key") as String? ?: System.getenv("GITHUB_TOKEN")
            }
        }
    }
}
