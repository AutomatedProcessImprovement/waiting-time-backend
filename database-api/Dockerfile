FROM python:3.9-slim

WORKDIR /app

RUN pip install gunicorn requests flask flask_cors psycopg2-binary pix-framework
RUN pip install --upgrade pip

COPY . /app

EXPOSE 5000

CMD ["gunicorn", "-w", "4", "app:app", "-b", "0.0.0.0:5000"]
