from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from app.db.models import Base
from config.settings import DATABASE_URL


replica_engine = create_engine(DATABASE_URL, pool_pre_ping=True)
ReplicaSession = sessionmaker(bind=replica_engine)
Base.metadata.create_all(replica_engine)
