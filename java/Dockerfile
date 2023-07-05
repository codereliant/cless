# syntax=docker/dockerfile:1

FROM eclipse-temurin:17-jdk-jammy as build

WORKDIR /app

COPY spring-petclinic/.mvn/ .mvn
COPY spring-petclinic/mvnw spring-petclinic/pom.xml ./

COPY spring-petclinic/src ./src
RUN ./mvnw package -DskipTests

FROM eclipse-temurin:17-jdk-jammy

COPY --from=build /app/target/*.jar /app/app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "/app/app.jar"]