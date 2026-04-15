@echo off
echo Starting data seeding process...

docker-compose run --rm seeder

echo Seeding completed successfully!
pause
