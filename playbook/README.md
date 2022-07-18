## Основная часть
1. Попробуйте запустить playbook на окружении из `test.yml`, зафиксируйте какое значение имеет факт `some_fact` для указанного хоста при выполнении playbook'a.

![изображение](https://user-images.githubusercontent.com/98019531/179554361-dc0ce393-5df3-41e9-9545-65cee6d274c1.png)

2. Найдите файл с переменными (group_vars) в котором задаётся найденное в первом пункте значение и поменяйте его на 'all default fact'.

![изображение](https://user-images.githubusercontent.com/98019531/179554973-5717425a-b01f-4f81-bf7c-4e958e6d6ddd.png)

3. Воспользуйтесь подготовленным (используется `docker`) или создайте собственное окружение для проведения дальнейших испытаний.

![изображение](https://user-images.githubusercontent.com/98019531/179561194-f47f385e-d444-4946-b480-a2d7453753c4.png)

4. Проведите запуск playbook на окружении из `prod.yml`. Зафиксируйте полученные значения `some_fact` для каждого из `managed host`.

![изображение](https://user-images.githubusercontent.com/98019531/179562394-1aa84d23-8331-4668-ad28-75b677250cde.png)

5. Добавьте факты в `group_vars` каждой из групп хостов так, чтобы для `some_fact` получились следующие значения: для `deb` - 'deb default fact', для `el` - 'el default fact'.
6.  Повторите запуск playbook на окружении `prod.yml`. Убедитесь, что выдаются корректные значения для всех хостов.

![изображение](https://user-images.githubusercontent.com/98019531/179562802-9cc31c27-74e6-4e65-adc5-2475a1c6c8f7.png)

7. При помощи `ansible-vault` зашифруйте факты в `group_vars/deb` и `group_vars/el` с паролем `netology`.

![изображение](https://user-images.githubusercontent.com/98019531/179563157-20ea396d-04ca-42a9-b3e4-b5d1b6c95e1c.png)

8. Запустите playbook на окружении `prod.yml`. При запуске `ansible` должен запросить у вас пароль. Убедитесь в работоспособности.

![изображение](https://user-images.githubusercontent.com/98019531/179563384-2393f334-8b5e-405b-8caa-455f09a18ad0.png)

9. Посмотрите при помощи `ansible-doc` список плагинов для подключения. Выберите подходящий для работы на `control node`.

![изображение](https://user-images.githubusercontent.com/98019531/179563735-201a0a81-c4a9-41d7-9dbd-4ec2adfb7faa.png)

10. В `prod.yml` добавьте новую группу хостов с именем  `local`, в ней разместите localhost с необходимым типом подключения.

![изображение](https://user-images.githubusercontent.com/98019531/179563948-4eb7fb53-ee57-4d09-98b6-acc84dd058c8.png)

11. Запустите playbook на окружении `prod.yml`. При запуске `ansible` должен запросить у вас пароль. Убедитесь что факты `some_fact` для каждого из хостов определены из верных `group_vars`.

![изображение](https://user-images.githubusercontent.com/98019531/179564086-00154a40-7b14-4e8c-8344-2ad6c4e48646.png)

12. Заполните `README.md` ответами на вопросы. Сделайте `git push` в ветку `master`. В ответе отправьте ссылку на ваш открытый репозиторий с изменённым `playbook` и заполненным `README.md`.


# Самоконтроль выполненения задания

1. Где расположен файл с `some_fact` из второго пункта задания?
```
    ./group_vars/all/examp.yaml
```
2. Какая команда нужна для запуска вашего `playbook` на окружении `test.yml`?
```
ansible-playbook -i inventory/test.yml site.yml
```
3. Какой командой можно зашифровать файл?
```
ansible-vault encrypt filename
```
4. Какой командой можно расшифровать файл?
```
ansible-vault decrypt --ask-vault-password filename
```
5. Можно ли посмотреть содержимое зашифрованного файла без команды расшифровки файла? Если можно, то как?
```
ansible-vault view filename
```
6. Как выглядит команда запуска `playbook`, если переменные зашифрованы?
```
ansible-playbook -i inventory/prod.yml site.yml --ask-vault-pass
```
7. Как называется модуль подключения к host на windows? 

winrm

8. Приведите полный текст команды для поиска информации в документации ansible для модуля подключений ssh
```
ansible-doc -t connection ssh
```
9. Какой параметр из модуля подключения `ssh` необходим для того, чтобы определить пользователя, под которым необходимо совершать подключение?
```
- remote_user
```
