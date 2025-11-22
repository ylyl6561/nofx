#!/usr/bin/env python3
"""
从生产环境同步配置数据到本地测试环境
"""

import os
import sys
import psycopg2
from psycopg2.extras import RealDictCursor
from datetime import datetime

# 需要同步的表和字段
SYNC_TABLES = {
    'system_ai_models': {
        'columns': ['id', 'name', 'provider', 'description', 'default_model_name', 'default_api_url', 'created_at'],
        'primary_key': 'id',
        'description': '系统AI模型模板'
    },
    'system_exchanges': {
        'columns': ['id', 'name', 'type', 'description', 'default_api_url', 'default_testnet_url', 'created_at'],
        'primary_key': 'id',
        'description': '系统交易所模板'
    },
    'system_config': {
        'columns': ['key', 'value', 'updated_at'],
        'primary_key': 'key',
        'description': '系统配置'
    },
    'beta_codes': {
        'columns': ['code', 'used', 'used_by', 'used_at', 'created_at'],
        'primary_key': 'code',
        'description': '内测码',
        'filter': 'used = false'  # 只同步未使用的
    }
}

def load_env_file(filename):
    """加载环境变量文件"""
    env_vars = {}
    try:
        with open(filename, 'r') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#') and '=' in line:
                    key, value = line.split('=', 1)
                    # 去除引号
                    value = value.strip().strip('"').strip("'")
                    env_vars[key] = value
    except FileNotFoundError:
        print(f"❌ 错误: 未找到文件 {filename}")
        sys.exit(1)
    return env_vars

def connect_db(database_url):
    """连接数据库"""
    try:
        conn = psycopg2.connect(database_url)
        return conn
    except Exception as e:
        print(f"❌ 数据库连接失败: {e}")
        sys.exit(1)

def export_table(conn, table_name, config):
    """从生产环境导出表数据"""
    cursor = conn.cursor(cursor_factory=RealDictCursor)
    
    columns = ', '.join(config['columns'])
    query = f"SELECT {columns} FROM {table_name}"
    
    if 'filter' in config:
        query += f" WHERE {config['filter']}"
    
    try:
        cursor.execute(query)
        rows = cursor.fetchall()
        cursor.close()
        return rows
    except Exception as e:
        print(f"   ❌ 导出失败: {e}")
        cursor.close()
        return []

def import_table(conn, table_name, config, rows):
    """导入数据到本地环境"""
    if not rows:
        print(f"   ⚠️  无数据可导入")
        return 0
    
    cursor = conn.cursor()
    imported = 0
    
    try:
        # 清空表
        cursor.execute(f"TRUNCATE TABLE {table_name} CASCADE")
        
        # 准备插入语句
        columns = config['columns']
        placeholders = ', '.join(['%s'] * len(columns))
        insert_query = f"INSERT INTO {table_name} ({', '.join(columns)}) VALUES ({placeholders})"
        
        # 批量插入
        for row in rows:
            values = [row[col] for col in columns]
            cursor.execute(insert_query, values)
            imported += 1
        
        conn.commit()
        cursor.close()
        return imported
    except Exception as e:
        conn.rollback()
        print(f"   ❌ 导入失败: {e}")
        cursor.close()
        return 0

def main():
    print("🔄 配置数据同步工具 (Python版)")
    print("=" * 60)
    print()
    
    # 获取项目根目录
    script_dir = os.path.dirname(os.path.abspath(__file__))
    project_root = os.path.dirname(script_dir)
    
    # 加载环境配置
    print("📋 加载环境配置...")
    prod_env = load_env_file(os.path.join(project_root, '.env.production'))
    local_env = load_env_file(os.path.join(project_root, '.env.local'))
    
    prod_db_url = prod_env.get('DATABASE_URL')
    local_db_url = local_env.get('DATABASE_URL')
    
    if not prod_db_url or not local_db_url:
        print("❌ 错误: 未找到 DATABASE_URL 配置")
        sys.exit(1)
    
    print(f"  生产: {prod_db_url[:50]}...")
    print(f"  本地: {local_db_url[:50]}...")
    print()
    
    # 连接数据库
    print("🔌 连接数据库...")
    prod_conn = connect_db(prod_db_url)
    local_conn = connect_db(local_db_url)
    print("  ✓ 连接成功")
    print()
    
    # 确认操作
    print("⚠️  这将覆盖本地数据库中的配置数据")
    response = input("是否继续? (y/n): ")
    if response.lower() != 'y':
        print("❌ 已取消")
        prod_conn.close()
        local_conn.close()
        sys.exit(0)
    print()
    
    # 同步数据
    print("=" * 60)
    print("📦 开始同步数据...")
    print("=" * 60)
    print()
    
    total_exported = 0
    total_imported = 0
    
    for table_name, config in SYNC_TABLES.items():
        print(f"📊 处理表: {table_name} ({config['description']})")
        
        # 导出
        print(f"   📤 从生产环境导出...")
        rows = export_table(prod_conn, table_name, config)
        print(f"   ✓ 导出 {len(rows)} 条记录")
        total_exported += len(rows)
        
        # 导入
        if rows:
            print(f"   📥 导入到本地环境...")
            imported = import_table(local_conn, table_name, config, rows)
            print(f"   ✓ 导入 {imported} 条记录")
            total_imported += imported
        
        print()
    
    # 关闭连接
    prod_conn.close()
    local_conn.close()
    
    # 总结
    print("=" * 60)
    print("✅ 同步完成！")
    print("=" * 60)
    print()
    print(f"📊 统计:")
    print(f"  导出: {total_exported} 条记录")
    print(f"  导入: {total_imported} 条记录")
    print()
    print("📋 已同步的表:")
    for table_name, config in SYNC_TABLES.items():
        print(f"  - {table_name}: {config['description']}")
    print()
    print("💡 提示:")
    print("  1. 用户数据和交易员配置未同步，避免冲突")
    print("  2. 决策日志和权益历史未同步")
    print("  3. 如需同步其他数据，请修改 SYNC_TABLES 配置")
    print()

if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n❌ 用户中断")
        sys.exit(1)
    except Exception as e:
        print(f"\n\n❌ 错误: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
