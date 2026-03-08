plugins {
    `java-library`
    `maven-publish`
}

group = "io.github.mystisql"
version = "1.0.0-SNAPSHOT"

repositories {
    mavenCentral()
}

java {
    sourceCompatibility = JavaVersion.VERSION_1_8
    targetCompatibility = JavaVersion.VERSION_1_8
    withJavadocJar()
    withSourcesJar()
}

dependencies {
    // JDBC API (provided by JDK, but needed for compilation)
    compileOnly("java.sql:java.sql-api:1.0")
    
    // HTTP Client
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    
    // JSON Serialization
    implementation("com.fasterxml.jackson.core:jackson-databind:2.16.1")
    implementation("com.fasterxml.jackson.datatype:jackson-datatype-jsr310:2.16.1")
    
    // Logging
    implementation("org.slf4j:slf4j-api:1.7.36")
    
    // Testing
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
