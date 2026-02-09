update users
set role = 'platform_admin'
where email = 'akpp91299@gmail.com';

select id, email, role
from users
where email = 'akpp91299@gmail.com';
