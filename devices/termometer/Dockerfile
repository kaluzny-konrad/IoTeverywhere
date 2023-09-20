# Użyj obrazu bazowego Pythona
FROM python:3

# Ustal katalog roboczy w kontenerze
WORKDIR /app

# Skopiuj plik z kodem źródłowym do katalogu roboczego w kontenerze
COPY termometer_simulator.py /app/termometer_simulator.py

# Zainstaluj niezbędne biblioteki
RUN pip install gpiozero

# Uruchom aplikację przy starcie kontenera
CMD [ "python", "./termometer_simulator.py" ]
