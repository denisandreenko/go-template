# Go service template

Шаблон сервиса. При клонировании указать имя директории для клонирования, которая имеет название сервиса.
После клонирования необходимо отредактировать файл Makefile.env и выполнить команду: make -f Makefile.env.
После этого станут доступны описанные ниже директивы.

## Сборка
```
Локально: make build
Docker: make docker_build
```
## Тесты
```
Запуск Unit tests в docker
make tests
```
## Развертывание
```
Поднять приложение и окружение в докере: make docker_env_up
Погасить все: make docker_env_down
```
```
Поднять только окружение в докере: make local_env_up
Погасить все: make local_env_down
```

## Дистрибутив
```
Создает архив с конфигами и приложением
make tar
```
```
Создает docker образ и пытается запушить его в docker registry
с тегами версии и latest    
make build_image
```

## Метрики
```
Добавлена возможность просматривать метрики сервиса 
после его старта в докере:

Адресс - localhost:3000
Login/Password - admin/P@ssw0rd (изменить можно в /deployments/docker/grafana/config.monitoring)
```
