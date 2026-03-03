#include <stdio.h>
#include <stdlib.h>

// Headers de nuestros archivos
#include "nodo.h"
#include "PDB.h"
#include "IDA.h"

void impresion(FILE *exit, struct nodo *meta){
	if(meta->ant != NULL)
		impresion(exit, meta->ant);
	for (int i = 0; i < 12; i++)
		fprintf(exit, "%d ", meta->estado[i]);
	fprintf(exit, "\n");
}

int main(int argc, char* argv[]) {
	if (argc != 14) {
        printf("❌ Error: Debes ingresar exactamente 12 numeros y el nombre del archivo de salida.\n");
        printf("Uso correcto: %s n1 n2 n3 n4 n5 n6 n7 n8 n9 n10 n11 n12 salida.txt\nDonde cada n es un numero entre 1 y 12\n", argv[0]);
        return 1;
    }

	int estadoInicial[12];
	int repeticiones[12] = {0,0,0,0,0,0,0,0,0,0,0,0}; // Chequeo de numeros repetidos en la instancia inicial
	char *salida = argv[13];
	for (int i = 0; i < 12; i++) {
        estadoInicial[i] = atoi(argv[i + 1]); 
    }
	// Validacion de datos validos
    for (int i = 0; i < 12; i++) {
        if(estadoInicial[i] > 12 || estadoInicial[i] < 1){	// Que este entre 1 y 12
			printf("❌ Error: Debes ingresar numeros entre 1 y 12.\n");
			printf("Uso correcto: %s n1 n2 n3 n4 n5 n6 n7 n8 n9 n10 n11 n12 salida.txt\nDonde cada n es un numero entre 1 y 12\n", argv[0]);
			return 1;
		}
		// Que no haya repeticiones
		if(estadoInicial[i] == 1){ 	repeticiones[0]++;}	if(repeticiones[0] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 2){ 	repeticiones[1]++;}	if(repeticiones[1] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 3){ 	repeticiones[2]++;}	if(repeticiones[2] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 4){ 	repeticiones[3]++;}	if(repeticiones[3] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 5){ 	repeticiones[4]++;}	if(repeticiones[4] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 6){ 	repeticiones[5]++;}	if(repeticiones[5] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 7){ 	repeticiones[6]++;}	if(repeticiones[6] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 8){ 	repeticiones[7]++;}	if(repeticiones[7] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 9){ 	repeticiones[8]++;}	if(repeticiones[8] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 10){ repeticiones[9]++;}	if(repeticiones[9] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 11){ repeticiones[10]++;}if(repeticiones[10] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
		if(estadoInicial[i] == 12){ repeticiones[11]++;}if(repeticiones[11] > 1){printf("❌ Error: No puede haber numeros repetidos.\n"); return 1;}
    }

	// PDBs
	const char *PDB1 = "PDB1.txt", *PDB2 = "PDB2.txt";

	// Arreglos que contienen las claves para los diferentes estados
	unsigned char *visitados1 = malloc(19958400 * sizeof(unsigned char));
	unsigned char *visitados2 = malloc(19958400 * sizeof(unsigned char));
    inicializarPDB(visitados1);
    inicializarPDB(visitados2);


	int info1[8], info2[8];// 	Para guardar informacion inicial de la particion

	// Llenado de PDB
	creacionPDB(PDB1, visitados1, info1);
	creacionPDB(PDB2, visitados2, info2);


	// Ejecutamos IDA*
	struct solYlim *resultado = IDA(estadoInicial, visitados1, info1, visitados2, info2);
	
	if(resultado == NULL){
		printf("❌ Error: No alcanzo el resultado.\n");
		return 1;
	}else{
		FILE *exit = fopen(salida, "w");
		if(exit != NULL){
			impresion(exit, resultado->nodo);
			fclose(exit);
			printf("(I)  La longitud del camino desde el estado inicial hasta la meta: %d\n",resultado->nodo->distanciaPadre);
			printf("(II) El numero de estados generados (evaluados): %lld\n", resultado->edoEvaluados);
			printf("(II) El numero de estados generados (explorados): %lld\n", resultado->edoExplorados);
		}else
			printf("❌ Error: No se pudo crear el archivo.\n");
	}
	// Documento para explicar heuristica
	free(visitados1);
	free(visitados2);
    return 0;
}