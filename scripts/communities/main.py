import json
import os
import mysql.connector


def get_inserts(root):
    communities_queue = [(root, None)]
    inserts = []
    while len(communities_queue) > 0:
        sub_communities, parent_id = communities_queue.pop()
        for sub_community_name, sub_community_children in sub_communities.items():
            community_id = len(inserts) + 1
            inserts.append((community_id, sub_community_name, parent_id))
            communities_queue.append((sub_community_children, community_id))
    return inserts


def bulk_insert_communities(inserts):
    with mysql.connector.connect(user=os.getenv("DB_USER"), password=os.getenv("DB_PASS"),
                                 host=os.getenv("DB_HOST"),
                                 database='next-dorm') as cnx:
        with cnx.cursor() as cursor:
            cursor.executemany("""
            INSERT INTO community (id, name, parent_id)
                VALUES (%s, %s, %s)
            """, inserts)
        cnx.commit()


if __name__ == '__main__':
    script_path = os.path.dirname(os.path.realpath(__file__))
    with open(f"{script_path}/communities.json") as f:
        communities = json.load(f)
    if communities is None:
        raise ValueError("an error occurred while loading in the communities")
    communities = get_inserts(communities)
    print(f"Inserting {len(communities)}")
    bulk_insert_communities(communities)
    print(f"Inserted {len(communities)} communities")
